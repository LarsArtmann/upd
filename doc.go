// Package upd upgrades NPM package dependencies in package.json files
// while preserving formatting. It is usable both as a library and via the
// upd command-line tool.
//
// The typical library flow is:
//
//	cfg := upd.DefaultConfig()
//	cfg.File = "package.json"
//	cfg.Nop = true // dry-run
//
//	pkg, err := upd.ReadPackageFile(cfg.File)
//	if err != nil { return err }
//
//	manifest := upd.BuildManifest(pkg, pkg.GetUpdArgs())
//	engine := upd.NewEngine(cfg)
//	results := engine.FetchAll(ctx, manifest.ToCheck())
//	updates, errs := engine.ApplyUpdates(manifest, results, pkg)
//	if updates > 0 && !cfg.Nop {
//	    _ = pkg.Write(cfg.File)
//	}
package upd
