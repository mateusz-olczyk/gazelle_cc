# Manual debug with testdata

Add a debug configuration for a test scenario from `language/cc/testdata/`. You may pass one argument.

**Argument:**
- **Directory name** (e.g. `cc_search`, `cc_flat_namespace`) – name of a directory under `language/cc/testdata/`. Use this to set up a temporary workspace and a debug configuration.
- **`cleanup`** – run after the user has finished debugging to remove the temporary directory and the added launch configuration.

---

## When the user provides a directory name (e.g. `cc_search`)

Do the following in order.

1. **Resolve paths**
   - Source: `language/cc/testdata/<directory>/`
   - Ensure this path exists; if not, tell the user and stop.

2. **Create temporary directory and copy**
   - Create a temporary directory: `TMPDIR=$(mktemp -d)` (or equivalent so you capture the path).
   - Recursively copy the contents of `language/cc/testdata/<directory>/` into that temporary directory (e.g. `cp -R` or `rsync` so the *contents* are copied, not the directory itself).

3. **Transform files inside the temporary directory**
   - Remove all `*.out` files under the temporary directory.
   - Rename all `*.in` files to `*.bazel` (e.g. `BUILD.in` → `BUILD.bazel`) under the temporary directory.

4. **Add a launch configuration to `.vscode/launch.json`**
   - If `language/cc/testdata/<directory>/arguments.txt` exists, read its lines (trimmed) and set **args** in the configuration to a JSON array of those lines. Otherwise omit the **args** key.
   - Insert a **new** configuration into the `configurations` array using this template. Replace `<TEMP_DIR_ABSOLUTE_PATH>` with the absolute path of the temporary directory (use the same value for both **name** and **cwd**). Replace `<directory>` in the configuration **name** with the directory name provided by the user. Replace `"args"` with the array of lines from `arguments.txt` when that file exists (e.g. `["-r=false", "-index=lazy", "use"]`), or remove the `"args"` key entirely when it does not. Preserve existing formatting, comments, and the rest of `launch.json`.

   ```json
   {
     "name": "manual-debug (language/cc/testdata/<directory>) (<TEMP_DIR_ABSOLUTE_PATH>)",
     "type": "go",
     "request": "launch",
     "mode": "exec",
     "program": "${workspaceFolder}/bazel-bin/language/cc/testdata/gazelle_for_testing_/gazelle_for_testing",
     "preLaunchTask": "bazel: build debuggable gazelle_cc for testing",
     "cwd": "<TEMP_DIR_ABSOLUTE_PATH>",
     "args": [],
     "substitutePath": [
       { "from": "${workspaceFolder}/bazel-gazelle_cc/external/gazelle+", "to": "external/gazelle+" },
       { "from": "${workspaceFolder}", "to": "" }
     ]
   }
   ```

5. **Instruct the user**
   - Tell the user to start the **Run and Debug** session and select the configuration whose name is **"manual-debug (language/cc/testdata/<directory>) (<TEMP_DIR_ABSOLUTE_PATH>)"** (with the actual directory and temp path you used).
   - Tell them: when they are done debugging, run this command again with the argument **`cleanup`** to remove the temporary directory(ies) and the added configuration(s).

---

## When the user provides `cleanup`

Do the following in order.

1. **Discover configurations to clean**
   - In `.vscode/launch.json`, find all configurations whose **name** matches the pattern `manual-debug (language/cc/testdata/...) (...)`, i.e. starts with `manual-debug (language/cc/testdata/` and contains a second parenthesized segment. Extract the path from that second parenthesis (the temporary directory path) for each matching configuration. If none match, tell the user there is nothing to clean up and stop.

2. **Remove temporary directories**
   - For each path extracted above, delete that directory if it exists (e.g. `rm -rf <path>`). Confirm existence before deleting.

3. **Remove launch configurations**
   - In `.vscode/launch.json`, remove every configuration whose **name** matched the pattern above. Leave all other configurations and the file structure unchanged.

4. **Confirm**
   - Tell the user that cleanup is done (how many configs and directories were removed).

---

## If no argument is provided

Ask the user for the directory name under `language/cc/testdata/` (e.g. `cc_search`) or for `cleanup` if they finished debugging.
