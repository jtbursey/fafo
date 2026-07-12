PROJECTNAME := fafo
PREFIX := fafo-
BINDIR := bin/

.DEFAULT_GOAL := $(PROJECTNAME)

.PHONY: all $(PROJECTNAME) configure clean

all: $(PROJECTNAME)

$(PROJECTNAME): configure
	go build -o ./$(BINDIR)$(PROJECTNAME) $(PROJECTNAME)/$(PROJECTNAME)

configure:
	go build -o ./$(BINDIR)configure $(PROJECTNAME)/configure

clean:
	$(RM) -r $(BINDIR)
