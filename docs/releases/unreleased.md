### Added

- Added `arcli mqtt` commands for health and aggregate stats, plus MQTT subscription list/show/create/update/delete, lifecycle actions, and per-subscription stats. List output supports table, JSON, and CSV; object output supports table and JSON. Create/update config is read from JSON files or stdin so credentials do not need to appear in process arguments; private file permissions are required when credentials are present, and subscription output omits usernames and broker URL userinfo. Delete, stop, pause, and restart prompt for confirmation. Closes [arcli#35](https://github.com/Basekick-Labs/arcli/issues/35).

- `arcli license status` inspects the public `/health` license block on standalone and clustered Arc deployments, with table and JSON output. Reports tier, server status, source, and optional expiry, days remaining, and site-license state. Expiry uses UTC RFC3339 with fractional seconds when provided.
- Absent/null license blocks succeed as normal OSS behavior and produce `{"configured":false,"license":null}` in JSON. Present licenses produce `{"configured":true,"license":{...}}`, preserving active, expired, invalid, unlicensed, and unknown server statuses. Optional absent/null fields are omitted; explicit zero days and false site-license values are retained.
- All successfully read license states exit 0. Configuration, request, decoding, and output failures exit nonzero. No token or database header is sent to `/health`, and no authentication verification is performed. Normal connection selection, timeout, TLS verification, and identity opt-out settings apply.

License installation and key/file configuration remain server responsibilities. Existing `ping -o json` license pass-through is unchanged. Closes [arcli#34](https://github.com/Basekick-Labs/arcli/issues/34).

- Preserve non-secret MQTT CSV field values separately from terminal sanitization and credential redaction. Subscription lists include created_at and updated_at timestamps. Table and CSV timestamps retain nanoseconds in UTC RFC3339Nano; JSON retains timestamp precision. Missing routes preserve the original HTTP error rather than implying MQTT is disabled.
