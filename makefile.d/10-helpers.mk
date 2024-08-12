####
## Detect OS and architecture
####

OSTYPE :=
ARCHTYPE :=

ifeq ($(OS),Windows_NT)
	OSTYPE = windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		ARCHTYPE = amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		ARCHTYPE = x86
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OSTYPE = linux
	endif
	ifeq ($(UNAME_S),Darwin)
		OSTYPE = darwin
	endif

	UNAME_P := $(shell uname -m)
	ifeq ($(UNAME_P),x86_64)
		ARCHTYPE = amd64
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		ARCHTYPE = x86
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		ARCHTYPE = arm64
	endif
endif
