# Pick my postcode

[![Build and test](https://github.com/andrew-field/pickmypostcode/actions/workflows/build-test.yml/badge.svg)](https://github.com/andrew-field/pickmypostcode/actions/workflows/build-test.yml)
[![Lint Code Base](https://github.com/andrew-field/pickmypostcode/actions/workflows/linter.yml/badge.svg)](https://github.com/andrew-field/pickmypostcode/actions/workflows/linter.yml)

2. If repository is private, remove the go report card workflow.
3. Add CodeQL and Dependabot workflow from Settings -> Code security and analysis (Package ecosystem should be "gomod"). Can't add CodeQL if the repository is private.
4. Update branch protection rule to require status checks, the branch is up to date and commits require a signature. Can't add these rules if the repository is private.
