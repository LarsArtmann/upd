// Package upd upgrades NPM package dependencies in package.json files
// while preserving formatting. It is usable both as a library and via the
// upd command-line tool.
//
// The typical library flow is:
//
//	cfg := upd.DefaultConfig()
//	cfg.File = "package.json"
//	cfg.Registry = "https://registry.npmjs.org"
//	cfg.Timeout = 20 * time.Second
//	cfg.Retries = 3
//	cfg.Nop = true // dry-run
//
//	pkg, err := upd.ReadPackageFile(cfg.File)
//	if err != nil { return err }
//
//	args, err := pkg.GetUpdArgs()
//	if err != nil { return err }
//
//	manifest, warnings := upd.BuildManifest(pkg, args, false)
//	engine := upd.NewEngine(cfg)
//	results := engine.FetchAll(ctx, manifest.ToCheck())
//	updates, errs := engine.ApplyUpdates(manifest, results, pkg)
//	if updates > 0 && !cfg.Nop {
//	    _ = pkg.Write(cfg.File)
//	}
package upd
