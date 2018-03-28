COMPILER=pdflatex
TEX=$(wildcard *.tex)
PROGS = $(patsubst %.tex,%,$(TEX))
TARGETS = $(patsubst %, %.pdf, $(PROGS))

all: $(TARGETS)

%.pdf: %.tex
	$(COMPILER) $<
	gnome-open $@
	rm -f *.aux *.log

clean:
	rm -f $(TARGETS)
