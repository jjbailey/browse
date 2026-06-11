# browse

**browse**: a multi-file pager with recursive navigation

**browse** is an interactive pager for navigating _sets of files_, not just
viewing one file at a time.

Unlike traditional pagers, **browse** lets you temporarily switch to a new set
of files and then return exactly where you left off. This recursive browsing
model makes it natural to explore logs, search results, source trees, and
related files without losing context.

**Think:** `less` + `pushd/popd` for files.

## Why Use Browse?

- Explore multiple files as one workflow.
- Drill into a new file set and return to your previous place.
- Browse command output from pipelines as if it were a file.
- Use keyboard-driven navigation, search, shell commands, and history.
- Keep your context while investigating logs, source, or generated results.

## Features

### Navigation Features

- Forward and reverse paging.
- Continuous scrolling in both directions.
- Horizontal scrolling for wide lines.
- Jump to line numbers.
- Mark pages and jump back to them.
- Follow and tail modes for changing files.

### Search and Exploration

- Forward and reverse regex and fixed-string search.
- Case-sensitive and case-insensitive search.
- Pattern highlighting.
- Search pattern history.
- Run `grep` on the current file in a nested browse session.

### Multi-File Workflow

- Browse multiple files from the command line.
- Browse standard input as a temporary file.
- Open nested file sets with `B`.
- Return from nested file sets with `x` or `X`.
- Rewind the active file list with `Ctrl+R`.
- Show the current remaining file list with `a`.

### Convenience

- File and directory completion.
- Shell escape with command completion.
- Persistent file, directory, search, and shell histories.
- Session saving and restoration.
- Run `fmt -s` on the current file in a nested browse session.
- Built-in help screen.

## Continuous Scrolling and Following

**browse** supports continuous scrolling and file following:

- Continuous scroll moves up or down until you stop it. Use `d` to scroll
  toward EOF and `u` to scroll toward SOF.
- When continuous scroll reaches EOF, **browse** enters follow mode and displays
  new lines as they are appended to the file.
- The tail command jumps to EOF and follows new output from there.
- The cursor shows whether follow mode is active. In follow mode, the cursor is
  in the lower left corner. Otherwise, it is in the upper left corner.

## Usage

Browse one or more files:

```bash
browse file1.log file2.log file3.log
```

Browse shell-expanded globs:

```bash
browse *.go
```

Browse results from a pipeline:

```bash
grep -rl timeout /var/log | browse
```

Start with an initial search pattern:

```bash
browse -p ERROR app.log
```

## Command Line Options

```bash
browse [OPTIONS] [FILE] [FILE...]
```

| Option                | Function                                      |
| --------------------- | --------------------------------------------- |
| `-f`, `--follow`      | Follow file changes while still browsing      |
| `-F`, `--tail`        | Follow file changes like `tail -f`            |
| `-i`, `--ignore-case` | Search ignores case                           |
| `-n`, `--numbers`     | Start with line numbers turned on             |
| `-p`, `--pattern`     | Initial search pattern                        |
| `-t`, `--title`       | Page title, default filename, blank for stdin |
| `-v`, `--version`     | Print browse version number                   |
| `-?`, `--help`        | Print browse command line options             |

## Keyboard Shortcuts

### Navigation Keys

| Key                           | Function                                    |
| ----------------------------- | ------------------------------------------- |
| `f`, `Page Down`, `Space`     | Page down toward EOF                        |
| `b`, `Page Up`                | Page up toward SOF                          |
| `Ctrl+F`, `Ctrl+D`, `z`       | Scroll half page down toward EOF            |
| `Ctrl+B`, `Ctrl+U`, `Z`       | Scroll half page up toward SOF              |
| `+`, `Right`, `Enter`         | Scroll one line toward EOF                  |
| `-`, `Left`                   | Scroll one line toward SOF                  |
| `d`, `Down`                   | Continuous scroll toward EOF, follow at EOF |
| `u`, `Up`                     | Continuous scroll toward SOF, stop at SOF   |
| `>`, `Tab`, `Ctrl+Right`      | Scroll right                                |
| `<`, `Backspace`, `Ctrl+Left` | Scroll left                                 |
| `^`                           | Scroll to column 1                          |
| `$`                           | Scroll to end of line                       |
| `e`, `End`                    | Jump to EOF, follow at EOF                  |
| `t`                           | Jump to EOF, tail at EOF                    |
| `j`                           | Jump to line number                         |
| `0`, `Home`                   | Jump to start of file, column 1             |
| `G`                           | Jump to end of file                         |
| `m`                           | Mark current page with number 1-9           |
| `1`-`9`                       | Jump to mark                                |

### Search

| Key | Function                                                           |
| --- | ------------------------------------------------------------------ |
| `/` | Regex search forward                                               |
| `?` | Regex search reverse                                               |
| `n` | Repeat search in current direction                                 |
| `N` | Repeat search in opposite direction                                |
| `i` | Toggle case-sensitive or case-insensitive search                   |
| `I` | Toggle regex or fixed-string search                                |
| `p` | Print current search pattern                                       |
| `P` | Clear search pattern                                               |
| `&` | Run `grep -nP` on current file for search pattern in a new session |

### Files, Lists, and Session Control

| Key      | Function                                        |
| -------- | ----------------------------------------------- |
| `B`      | Browse another file or file set                 |
| `R`      | Re-read the current file from disk              |
| `Ctrl+R` | Rewind the current browse list                  |
| `a`      | Print filenames in the current browse list      |
| `q`      | Quit current file, save session, continue list  |
| `Q`      | Quit current file without saving, continue list |
| `x`      | Exit current list, save session                 |
| `X`      | Exit current list without saving session        |
| `Ctrl+X` | Exit nested list, save session                  |
| `Ctrl+Y` | Exit nested list without saving session         |

### Miscellaneous

| Key                | Function                                      |
| ------------------ | --------------------------------------------- |
| `#`                | Toggle line numbers                           |
| `%`, `=`, `Ctrl+G` | Show file position                            |
| `!`                | Run a shell command                           |
| `F`                | Run `fmt -s` on current file in a new session |
| `c`                | Print current working directory               |
| `C`                | Change working directory                      |
| `h`                | Show the help screen                          |
| `H`                | Show the man page                             |

## Working With Files

### Opening File Sets

Press `B` to open a new file or file set. The prompt accepts one or more file
names, shell globs such as `*.go`, quoted filenames containing spaces, and
history entries. It also expands special symbols such as `%` for the current
file and `~` for your home directory.

When you open a file set with `B`, **browse** temporarily leaves the current
list. When the nested list is finished, **browse** automatically resumes the
previous list where you left off.

### Showing the Current List

Press `a` to show the current file and any remaining files in the active list.
If you started with:

```bash
browse file1 file2 file3
```

and are currently viewing `file2`, pressing `a` shows `file2` and `file3`.

### Re-Reading Files

Press `R` to re-read the current file from disk. This is useful when a file is
rewritten in place, replaced, truncated, or otherwise changed in a way that the
automatic file tracking did not fully capture.

For example:

```bash
mv log log.old
app > log
```

If the screen stops matching the file you expect, `R` is the manual escape
hatch: it reopens the original path and rebuilds the browse state from disk.

### Rewinding Lists

Press `Ctrl+R` to rewind the active browse list. This returns to the first file
in the current list, including a nested list opened with `B`, without rewinding
any parent list.

### Changing Directory

Press `C` to change the current working directory. The prompt accepts `~`, `-`,
and `~-`, quoted directory names, and `%` for the parent directory of the
current file. Directory completion includes entries from `CDPATH`.

## Symbol Expansions

Special symbols are expanded in file and directory prompts, and shell commands:

| Symbol | Expands To                                                        |
| ------ | ----------------------------------------------------------------- |
| `!`    | Last shell command                                                |
| `%`    | Current file name; parent directory of current file in `C` prompt |
| `&`    | Current search pattern                                            |
| `~`    | Home directory                                                    |

## Configuration and History

**browse** stores configuration and history in:

```text
~/.browse/
```

The session file is:

```text
~/.browse/browserc
```

It saves:

- Current file name.
- First line on the page.
- Search pattern.
- Marks.
- Page title.
- Search case-sensitivity mode.
- Fixed-string search mode.

History files are maintained for common workflows, behaving like Bash history:

- **Shell commands** (`!` key): Every shell command is remembered and can be
  recalled, edited, and rerun.
- **Directory history** (`C` key): Recently visited directories are recorded
  for fast navigation across large projects.
- **File history** (`B` prompt): Browsed files are saved as full pathnames and
  are available in current and future sessions.
- **Search patterns** (`/` and `?` prompts): Regex and text search patterns are
  saved so you can repeat or revisit common queries without retyping them.

History files:

- `~/.browse/browse_dirs` - directory history.
- `~/.browse/browse_files` - file browsing history.
- `~/.browse/browse_search` - search pattern history.
- `~/.browse/browse_shell` - shell command history.

## Limitations

- Xterm-specific behavior.
- Displayed lines are clipped to screen width, with horizontal scrolling
  available for wider lines.
- Long lines are internally capped at about 4K.
- Tabs are converted to spaces.
- Non-printable characters may display poorly.
- Terminal title handling may vary by environment.

## License

MIT License - see LICENSE file for details.
