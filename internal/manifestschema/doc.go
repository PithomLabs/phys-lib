// Package manifestschema pins the General Relativity manifest key-set
// contract and machine-enforces every phys-lib manifest's conformance to it.
//
// It is test-only conformance machinery: it reads manifest.json files and
// compares their top-level key sets against the GR precedent. It does not
// import any phys-* framework and does not touch a PhysMath artifact.
package manifestschema
