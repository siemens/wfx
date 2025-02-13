-- SPDX-FileCopyrightText: 2023 Siemens AG
--
-- SPDX-License-Identifier: Apache-2.0
--
-- Author: Michael Adler <michael.adler@siemens.com>

--[[
A Lua filter to rewrite (and fix) links for Hugo deployment.

--]]
-- can also use https://github.com/pandoc-ext/logging as a drop-in
-- local logging = require("logging")
local logging = {
	info = function(...)
		io.stderr:write("[INFO] ")
		io.stderr:write(...)
		io.stderr:write("\n")
	end,
	error = function(...)
		io.stderr:write("[ERROR] ")
		io.stderr:write(...)
		io.stderr:write("\n")
	end,
}

local link_rewrites = {
	-- src -> dest
	["../workflow/dau/README.md"] = "workflows/dau.md",
	["../CONTRIBUTING.md"] = "https://github.com/siemens/wfx/tree/main/CONTRIBUTING.md",
	["../contrib/remote-access/README.md"] = "https://github.com/siemens/wfx/tree/main/contrib/remote-access",
	["../contrib/remote-access/wfx.workflow.remote.access.yml"] = "https://github.com/siemens/wfx/tree/main/contrib/remote-access/wfx.workflow.remote.access.yml",
	["../contrib/config-deployment/wfx.workflow.config.deployment.yml"] = "https://github.com/siemens/wfx/tree/main/contrib/config-deployment/wfx.workflow.config.deployment.yml",
	["../contrib/config-deployment/README.md"] = "https://github.com/siemens/wfx/tree/main/contrib/config-deployment",
	["../contrib/config-deployment/client/worker.go"] = "https://github.com/siemens/wfx/tree/main/contrib/config-deployment/client/worker.go",
	["../fbs"] = "https://github.com/siemens/wfx/tree/main/fbs",
	["../example/plugin"] = "https://github.com/siemens/wfx/tree/main/example/plugin",
}

local function find_git_root()
	local p = io.popen("git rev-parse --show-toplevel", "r")
	if not p then
		error("please install git")
	end
	local result = p:read("*l")
	p:close()
	return result
end

local function copy_file(from, to)
	logging.info("Copying file: ", from, " -> ", to)
	local f = io.open(from, "rb")
	if not f then
		error(string.format("ERROR: source %s does not exist", from))
	end

	local block = f:read(4096)
	if not block then
		logging.info("Skipping ", from, " because it is not a readable file")
		f:close()
		return
	end


	local t = io.open(to, "wb")
	if not t then
		error(string.format("ERROR: Cannot create destination %s", to))
	end

	repeat
		t:write(block)
		block = f:read(4096)
	until block == nil

	f:close()
	t:close()
end

local git_root = find_git_root()

--- Filter function for links
local function link(el)
	-- link destination
	local target = el.target
	local rewrite_target = link_rewrites[target]
	if rewrite_target then
		el.target = rewrite_target
		return el
	end

	if string.match(target, "^http") then
		-- link is external, no need to rewrite anything
		return nil
	end

	local fname = pandoc.path.filename(target)
	local is_markdown = string.match(fname, "%.md$") or string.match(fname, "%.md#") or string.match(fname, "^#")
	if not is_markdown then
		-- target is not a markdown document; copy it to /static.
		-- ASSUMPTION: working directory is set to the basedir of fname;
		-- this can only be done externally, i.e. must be done (somehow) *before*
		-- invoking pandoc
		local dest = string.format("%s/hugo/static/%s", git_root, fname)
		copy_file(target, dest)
		local new_target = "/" .. fname
		logging.info("Rewriting link: ", el.target, " -> ", new_target)
		el.target = new_target
	end
	return el
end

local function image(el)
	local fname = pandoc.path.filename(el.src)
	local new_target = "/" .. fname
	-- copying step happens in link() function
	el.src = new_target
	return el
end

return {
	{ Link = link },
	{ Image = image },
}

-- vim: ts=4 sw=4 noexpandtab
