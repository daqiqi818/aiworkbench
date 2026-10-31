# OpenMontage Upstream Record

This directory is a clean source snapshot of OpenMontage.

- Upstream repository: `https://github.com/dajiang7788/OpenMontage.git`
- Pinned commit: `9590f56`
- Imported source license: GNU Affero General Public License v3.0 or later
- Initial supported pipeline: `pipeline_defs/cinematic.yaml`

The import is produced from `git archive` at the pinned commit. It deliberately excludes Git
metadata, local environment files, credentials, Python and Node dependency directories, generated
projects, caches, bytecode, and other workstation data. Local changes to the vendored tree must be
recorded in this file and in the public source distribution.

The native OpenMontage project workspace is not stored in this source directory. The runner writes
one scoped workspace per production under its configured persistent workspace root.
