Validator for sigs.yaml
=======================

Tool to automatically validate content from [sigs.yaml](../../../sigs.yaml)

Features:
* Check/remove unreachable owners links

Update sigs.yaml with result from validator

Usage
-----

```bash
go run validators/cmd/sigs/sigs-validator.go --sigs_file_path=./sigs.yaml --dry-run=false
```
