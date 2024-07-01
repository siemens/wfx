module github.com/siemens/wfx/example/plugin

replace github.com/siemens/wfx => ../..

go 1.22.3

toolchain go1.22.4

require github.com/siemens/wfx v0.3.0

require github.com/google/flatbuffers v24.3.25+incompatible // indirect
