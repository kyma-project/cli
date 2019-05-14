---
name: Decision record
about: Decision record
---

Created on {YYYY-MM-DD} by {name and surname (@Github username)}.

## Decision log

| Name | Description |
|-----------------------|------------------------------------------------------------------------------------|
| Title | {Provide a brief summary of the decision.} |
| Ultimate decision maker(s) | {Specify who takes the final decision if the SIG/WG cannot reach the consensus.} |
| Due date | {Specify the date by which the SIG or WG members need to make the decision. Use the `YYYY-MM-DD` date format.} |
| Status | {The status of the document can be `Accepted`, `Declined` or `Proposed` (when it is waiting for decision). This section should contain one of these three words followed by the date on which the status of the document is agreed on. Follow the `YYYY-MM-DD` format for the date. For example: Proposed on 2018-03-20. Accepted on 2018-03-23. Add the new status change when it happens. Do not overwrite the previous status change.}|
| Input provider(s) | {Include those who influence the decision and, if possible, those whom the decision affects.} |
| Group(s) affected by the decision | {Specify whom the decision affects.} |
| Decision type | {Type in `Binary`, `Choice`, or `Prioritization`. The `Binary` type refers to the  yes/no decisions, the `Choice` type means that the decision involves choosing between many possibilities, such as a name for a new product, and the `Prioritization` type involves ranking a number of options, such as choosing the next five features to build out of one hundred possible options.} |
| Earliest date to revisit the decision | {Specify the date from which the SIG/WG members can raise an issue to revisit the decision. Use the `YYYY-MM-DD` date format. The date does not apply to ultimate decision makers who can raise an issue to revisit the decision anytime.} |
| Affected decisions | {Specify the id of the decision issue or a link to the previous decision record which is affected by this decision. Use the `#{issueid}\|{decision-record-URL}(replaces\|extends\|depends on)` format. For example, write `#265(replaces)` or `#278(depends on)` which means that the decision you propose replaces the issue 265 or depends on the issue 278. Specify as many references as possible and separate them with a comma. Write `None` if no other decision is affected.}|

## Context

<!-- Briefly describe what the decision record (DR) is about. 
Explain the factors for the decision, what are the forces at play, and the reasons why the discussed solution is needed. 
Remember that this document should be relatively short and concise. If necessary, provide relevant links for more details.
If the decision concerns more solutions, mark them with separate subsections. Use H3 for the subsection headings.  -->

## Decision

<!--Avoid using personal constructions such as "we." Use impersonal forms instead. 
For example, `The decision is to...`. If it is necessary to indicate the subject, use `SIG/WG members` instead of "we." -->

## Consequences

<!-- Briefly explain the consequences of this decision for the Kyma project. 
Include both the advantages and disadvantages of the discussed solution.
-->
