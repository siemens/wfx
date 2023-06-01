//go:generate find . -not -name generate.go -and -not -name configure_workflow_executor.go -and -not -path "./ent/*" -type f -delete

//go:generate just -d . --justfile ../spec/justfile generate wfx.swagger.yml
//go:generate swagger generate model --copyright-file=../spec/spdx.txt --target=. --model-package=model --spec=wfx.swagger.yml

//go:generate swagger generate server --copyright-file=../spec/spdx.txt --target=northbound --spec=wfx.swagger.yml --exclude-main --skip-models --model-package=model --existing-models=github.com/siemens/wfx/generated/model --flag-strategy=pflag --tags=northbound
//go:generate rm -f northbound/restapi/server.go

//go:generate swagger generate server --copyright-file=../spec/spdx.txt --target=southbound --spec=wfx.swagger.yml --exclude-main --skip-models --model-package=model --existing-models=github.com/siemens/wfx/generated/model --flag-strategy=flag --tags=southbound
//go:generate rm -f southbound/restapi/server.go

//go:generate swagger generate client --copyright-file=../spec/spdx.txt --target=. --model-package=model --spec=wfx.swagger.yml --skip-models --existing-models=github.com/siemens/wfx/generated/model

package generated

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */
