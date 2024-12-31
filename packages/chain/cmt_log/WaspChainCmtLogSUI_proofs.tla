---- MODULE WaspChainCmtLogSUI_proofs ------------------------------------------
EXTENDS WaspChainCmtLogSUI, TLAPS, SequenceTheorems, SequencesExtTheorems

LEMMA
    ASSUME NEW S, NEW x \in S
    PROVE <<x>> \in Seq(S)
PROOF
    BY ElementOfSeq


LEMMA
    ASSUME NEW S, IsFiniteSet(S)
    PROVE SetToSeq(S) \in Seq(S)


THEOREM (* Incomplete, only a draft. *)
    Spec => []TypeOK
PROOF
    <1>1. Init => TypeOK
        BY ElementOfSeq, HeadTailProperties DEF Init, TypeOK, SomeAoOrder
    <1>2. TypeOK /\ [Next]_vars => TypeOK'
        <2> SUFFICES ASSUME TypeOK, Next PROVE TypeOK' BY DEF vars, TypeOK
        <2>01. CASE \E n \in Node : AdvanceNodeView(n)
        <2>02. CASE \E n \in Node : AdvanceLastLI(n)
        <2>03. CASE \E n \in Node : ProposeOnMinLI(n)
        <2>04. CASE \E n \in Node : ProposeOnPrevLI(n)
        <2>05. CASE \E n \in Node : ProposeOnPrevLIBot(n)
        <2>06. CASE \E n \in Node : ProposeOnPrevLIOut(n)
        <2>07. CASE \E n \in Node : ProposeOnTimeout(n)
        <2>08. CASE \E n \in Node : ObtainL2Output(n)
        <2>09. CASE \E n \in Node : PublishL2Output(n)
        <2>10. CASE \E n \in Node : DropOldInstances(n)
        <2>11. CASE \E n \in Node : Reboot(n)
        <2>12. CASE DecideOnL2
        <2> QED BY <2>01, <2>02, <2>03, <2>04, <2>05, <2>06, <2>07, <2>08, <2>09, <2>10, <2>11, <2>12
            DEF Next, NextFair, NextFail
    <1>3. QED BY <1>1, <1>2, PTL DEF Spec


================================================================================
