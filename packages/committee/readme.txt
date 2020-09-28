consensus states

Consensus operator
- INIT immediately after activation
- NEW_STATE immediately after transition to the new state
- NEW_LEADER immediately after leader change
non-leader states
- NOTIFICATIONS_SENT for non-leader immediately after notifications sent to the leader
- VM_RUN_STARTED after was received request for calculation from the leader which is assumed
- RESULT_SENT result was sent to the leader
- CALCULATION_EVIDENCE_received from the leader
- CALCULATION_EVIDENCE_received from the tangle (not confirmed yet)
leader-states
- RUN_CALCULATIONS sent to peers and own calculation started
- RESULT-finalize (sent to the tangle and peers)