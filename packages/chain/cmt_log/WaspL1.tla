---- MODULE WaspL1 ----
EXTENDS Sequences
CONSTANT AOs

VARIABLE chain
VARIABLE posted

TX == [inp: AOs, out: AOs]
tx(inp, out) == [inp |-> inp, out |-> out]
txOnHead(tx) == tx.inp = Head(chain)

Init ==
    /\ chain = <<0>> \* TODO: ...
    /\ posted = {}

PostTX(inp, out) ==
    LET t == tx(inp, out)
    IN  /\ posted' = posted \cup (IF txOnHead(t) THEN {t} ELSE {})  \* Reject the TX immediately, if it is not on the head.
        /\ UNCHANGED chain                                          \* Chain is updated async, on confirm.

Confirm ==
    \E p \in posted :
        /\ p.inp = chain[0]                           \* We consume the unspent output.
        /\ chain' = <<p.out>> \o chain                \* append the AO to the chain.
        /\ posted' = {pp \in posted : pp.inp = o.inp} \* Consume the TX and reject all the other TXes consuming the same input.

====