# TLC_OPTS = -workers 16 -coverage 1
TLC_OPTS = -workers 16

WaspChainAccessNodesV4 = \
	WaspChainAccessNodesV4_N2C1L3R0.out \
	WaspChainAccessNodesV4_N2C1L3R1.out \
	WaspChainAccessNodesV4_N2C1L4R0.out \
	WaspChainAccessNodesV4_N2C1L5R0.out

all: $(WaspChainAccessNodesV4)

$(WaspChainAccessNodesV4): %.out : %.cfg WaspChainAccessNodesV4.tla
	tlc $(TLC_OPTS) -config $< $(word 2,$^) > $@
CLEAN_FILES += $(WaspChainAccessNodesV4)

clean:
	rm -rf $(CLEAN_FILES)

.PHONY: all clean
