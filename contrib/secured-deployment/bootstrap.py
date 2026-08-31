#!/usr/bin/env python3
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

import argparse
import base64
import datetime
import hashlib
import logging
import os
from pathlib import Path
import secrets

try:
    from cryptography import x509
    from cryptography.hazmat.primitives import hashes, serialization
    from cryptography.hazmat.primitives.asymmetric import padding, rsa
    from cryptography.x509.oid import NameOID
except ImportError:
    raise ImportError(
        "Missing python3-cryptography package."
    )

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

parser = argparse.ArgumentParser()
parser.add_argument("--local", action="store_true")
parser.add_argument("--state", type=Path, default=Path(__file__).parent / "state")
args = parser.parse_args()

os.umask(0o077)
state = args.state.resolve()
logger.info("Initializing state directory at %s", state)
state.mkdir(parents=True, exist_ok=True)
(state / "redis").mkdir(exist_ok=True)


def random_base64(length):
    return base64.b64encode(os.urandom(length)).decode().translate(str.maketrans("+/", "-_"))


def write_private(path, content):
    logger.debug("Writing private file: %s", path)
    if isinstance(content, bytes):
        path.write_bytes(content)
    else:
        path.write_text(content)
    path.chmod(0o600)


def password_digest(password):
    salt_bytes = os.urandom(16)
    salt = base64.b64encode(salt_bytes).decode().rstrip("=").replace("+", ".")
    digest = base64.b64encode(hashlib.pbkdf2_hmac(
        "sha512", password.encode(), salt_bytes, 310000
    )).decode().rstrip("=").replace("+", ".")
    return f"$pbkdf2-sha512$310000${salt}${digest}"


secrets_file = state / "secrets.env"
if not secrets_file.is_file():
    logger.info("Generating %s", secrets_file.name)
    write_private(secrets_file, "\n".join((
        f"AUTHELIA_SESSION_SECRET={secrets.token_hex(32)}",
        f"AUTHELIA_STORAGE_ENCRYPTION_KEY={secrets.token_hex(32)}",
        f"AUTHELIA_IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET={secrets.token_hex(32)}",
        f"AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET={secrets.token_hex(32)}",
        f"OAUTH2_PROXY_COOKIE_SECRET={random_base64(32)}",
        "",
    )))
else:
    logger.debug("%s already exists, skipping", secrets_file.name)

client_credentials = (
    "oauth2-client-secret",
    "wfxctl-client-secret",
)
for name in client_credentials:
    secret_path = state / name
    digest_path = state / f"{name}-digest"
    if not (secret_path.is_file() and digest_path.is_file()):
        logger.info("Generating secret for %s", name)
        secret_val = random_base64(32)
        write_private(secret_path, secret_val)
        write_private(digest_path, f"{password_digest(secret_val)}\n")
    else:
        logger.debug("Secret %s already exists, skipping", name)

# One Authelia user per device. The password is only needed for the interactive
# device-code enrollment; tokens are refreshed headlessly afterwards.
devices = (
    "wfx-client-1",
    "wfx-client-2",
)
device_password_digests = {}
for name in devices:
    password_path = state / f"{name}-password"
    digest_path = state / f"{name}-password-digest"
    if not (password_path.is_file() and digest_path.is_file()):
        logger.info("Generating password for %s", name)
        password = random_base64(24)
        write_private(password_path, password)
        write_private(digest_path, f"{password_digest(password)}\n")
    else:
        logger.debug("Password for %s already exists, skipping", name)
    device_password_digests[name] = digest_path.read_text().strip()

users_substitutions = [
    (f"@PASSWORD_DIGEST_{name.replace('-', '_').upper()}@", digest)
    for name, digest in device_password_digests.items()
]


def check_certs(certs_dir, cert_files):
    if not all((certs_dir / f).is_file() for f in cert_files):
        return False
    try:
        ca_cert = x509.load_pem_x509_certificate((certs_dir / "ca.pem").read_bytes())
        leaf_cert = x509.load_pem_x509_certificate((certs_dir / "cert.pem").read_bytes())
        now = datetime.datetime.now(datetime.timezone.utc)
        if not (ca_cert.not_valid_before_utc <= now <= ca_cert.not_valid_after_utc):
            return False
        if not (leaf_cert.not_valid_before_utc <= now <= leaf_cert.not_valid_after_utc):
            return False
        ca_pubkey = ca_cert.public_key()
        ca_pubkey.verify(
            leaf_cert.signature,
            leaf_cert.tbs_certificate_bytes,
            padding.PKCS1v15(),
            leaf_cert.signature_hash_algorithm,
        )
        return True
    except Exception:
        return False


def generate_rsa_key():
    return rsa.generate_private_key(public_exponent=65537, key_size=2048)


def key_to_pem(key):
    return key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.TraditionalOpenSSL,
        encryption_algorithm=serialization.NoEncryption(),
    )


def cert_to_pem(cert):
    return cert.public_bytes(serialization.Encoding.PEM)


certs = state / "certs"
certificate_files = ("ca.pem", "ca-key.pem", "cert.pem", "key.pem", "oidc-key.pem", "oidc-cert.pem")
ca_valid = check_certs(certs, certificate_files)

if not ca_valid:
    logger.info("Generating TLS certificates in %s", certs)
    certs.mkdir(exist_ok=True)
    now = datetime.datetime.now(datetime.timezone.utc)
    expiry = now + datetime.timedelta(days=3650)

    # CA Certificate
    ca_key = generate_rsa_key()
    ca_name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "wfx local CA")])
    ca_cert = (
        x509.CertificateBuilder()
        .subject_name(ca_name)
        .issuer_name(ca_name)
        .public_key(ca_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now)
        .not_valid_after(expiry)
        .add_extension(x509.BasicConstraints(ca=True, path_length=None), critical=True)
        .add_extension(
            x509.KeyUsage(
                digital_signature=False,
                content_commitment=False,
                key_encipherment=False,
                data_encipherment=False,
                key_agreement=False,
                key_cert_sign=True,
                crl_sign=True,
                encipher_only=False,
                decipher_only=False,
            ),
            critical=True,
        )
        .sign(ca_key, hashes.SHA256())
    )
    write_private(certs / "ca-key.pem", key_to_pem(ca_key))
    write_private(certs / "ca.pem", cert_to_pem(ca_cert))

    # Leaf Certificate
    leaf_key = generate_rsa_key()
    leaf_name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "wfx-mgmt.localhost")])
    leaf_cert = (
        x509.CertificateBuilder()
        .subject_name(leaf_name)
        .issuer_name(ca_cert.subject)
        .public_key(leaf_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now)
        .not_valid_after(expiry)
        .add_extension(
            x509.SubjectAlternativeName([
                x509.DNSName("wfx-mgmt.localhost"),
                x509.DNSName("wfx.localhost"),
                x509.DNSName("authelia.localhost"),
            ]),
            critical=False,
        )
        .sign(ca_key, hashes.SHA256())
    )
    write_private(certs / "key.pem", key_to_pem(leaf_key))
    write_private(certs / "cert.pem", cert_to_pem(leaf_cert))

    # OIDC Self-Signed Certificate
    oidc_key = generate_rsa_key()
    oidc_name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "oidc-signing.local")])
    oidc_cert = (
        x509.CertificateBuilder()
        .subject_name(oidc_name)
        .issuer_name(oidc_name)
        .public_key(oidc_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now)
        .not_valid_after(expiry)
        .sign(oidc_key, hashes.SHA256())
    )
    write_private(certs / "oidc-key.pem", key_to_pem(oidc_key))
    write_private(certs / "oidc-cert.pem", cert_to_pem(oidc_cert))
else:
    logger.debug("Certificates exist and are valid, skipping generation")

local = args.local
logger.info("Applying template configurations (local=%s)", local)
replacements = {
    "authelia/configuration.yml": ("authelia.yml", (
        ("@STATE@", str(state) if local else "/data"),
        ("tcp://127.0.0.1:9091", "tcp://127.0.0.1:9091" if local else "tcp://0.0.0.0:9091"),
        ("host: 127.0.0.1", "host: 127.0.0.1" if local else "host: redis"),
    )),
    "authelia/users.yml": ("users.yml", users_substitutions),
    "oauth2-proxy.cfg": ("oauth2-proxy.cfg", (
        ("@STATE@", str(state) if local else "/data"),
        ("127.0.0.1", "127.0.0.1" if local else "0.0.0.0"),
    )),
    "oauth2-proxy-alpha.yml": ("oauth2-proxy-alpha.yml", (
        ("@STATE@", str(state) if local else "/data"),
        ("127.0.0.1", "127.0.0.1" if local else "oauth2-proxy"),
    )),
}
source_root = Path(__file__).parent
for source, (destination, substitutions) in replacements.items():
    src_path = source_root / source
    dst_path = state / destination
    logger.debug("Rendering %s -> %s", src_path, dst_path)
    content = src_path.read_text()
    for old, new in substitutions:
        content = content.replace(old, new)
    dst_path.write_text(content)

logger.info("Setting file permissions")
for path in (
    "secrets.env", "users.yml",
    "oauth2-client-secret", "oauth2-client-secret-digest",
    "wfxctl-client-secret", "wfxctl-client-secret-digest",
    *(f"{name}-{suffix}" for name in devices for suffix in ("password", "password-digest")),
):
    (state / path).chmod(0o600)
for path in certs.glob("*key.pem"):
    path.chmod(0o600)

logger.info("Setup complete")
