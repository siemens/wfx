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

local function file_exists(fname)
	local f = io.open(fname, "r")
	if not f then
		return false
	end
	f:close()
	return true
end

--- Filter function for links
local function link(el)
	local target = el.target
	-- we only care about links to markdown documents != self
	local is_markdown_foreign = not string.match(target, "^http")
		and (string.match(target, "%.md$") or string.match(target, "%.md#"))
	if is_markdown_foreign then
		-- strip fragment
		local idx = string.find(target, "#")
		if idx then
			target = string.sub(target, 1, idx - 1) -- remove the part starting with "#"
		end
		if not file_exists(target) then
			local new_target = string.format("../%s", target)
			if file_exists(new_target) then
				logging.info("fixed dead link:", target, "->", new_target)
				el.target = new_target
			else
				logging.error("unable to fix dead link:", target)
			end
		end
	end
	return el
end

return {
	{ Link = link },
}

-- vim: ts=4 sw=4 noexpandtab
