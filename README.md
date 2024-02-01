# gha-updates

**Stay up-to-date with the latest GitHub Actions versions in your workflows!**

This tool scans your GitHub workflow YAML files and alerts you to any available updates for the actions being used.

## Getting Started

1. **Build the tool:**

   ```bash
   go build
   ```

2. **Run the tool:**

   ```bash
   ./ghs-updates <list of GitHub workflow YAML files>
   ```

   Example:

   ```bash
   ./ghs-updates .github/workflows/ci.yml .github/workflows/release.yml
   ```

## Output

The tool will display a report indicating:

- **Actions with available updates:** For each action with a newer version, it will list:
    - The current version used in the workflow
    - The latest available version

## License

This project is licensed under the MIT License. See the LICENSE file for details.
