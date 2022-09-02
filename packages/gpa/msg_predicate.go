// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

//
// TODO: Create an event handler, that acts as a local predicate.
// It collects messages, and eventually output another message.
// Maybe that will allow to implement "wait until quorum, etc",
// scenarios in more declarative way. Especially this should help
// with the scenarios, where multiple events of different kinds
// are awaited before proceeding. Implementing this explicitly one
// has to add predicate checks in multiple places (where the source
// events can happen), that's error prone.
//
// This should be local, used inside of a gpa impl.
//
// CEP stands for complex event processing. Maybe name should be changed.
// Something with predicates maybe. `msg_predicate`?
//
// It can also provide some state representation, to show, what is missing
// for a predicate to become true.
//
