# Changelog

## 2026-10-23

/speckit-specify The library is pure Markdown to/from  subject-predicate-object triples codec. The markdown format is governed by https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md . Implement the core types to be used as a public interface between the client application and this library so that client operates with Markdown files as bag of triples. 

/speckit-plan Implement the markdown to triple codec as pure function that operates with `io.Reader` and `[]byte`. Define core types to represent Markdown node as bag of triples. Ensure usage of `iter.Seq` as an itterator for the triples. Supports a sum-type for "object" so that application can dynamically handle a typed representation (make a research for the best design of type safety of this concept). Declare Node ID as own type rather any generic one. Implement CI/CD as github actions following the example: https://github.com/fogfish/arcnet-cli/tree/main/.github/workflows
