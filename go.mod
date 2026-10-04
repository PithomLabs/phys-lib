module github.com/PithomLabs/phys-lib

go 1.25.7

require github.com/PithomLabs/phys-math v0.0.0

// M-4 DEFERRED under D17: Phase 1 uses go.work for local multi-module
// development; standalone publication/version resolution is deferred.
// No replace directive is introduced solely to satisfy module tooling
// outside the workspace.
