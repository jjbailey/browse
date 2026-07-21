# browse

**browse** is an interactive pager for navigating collections of files rather
than only a single file at a time.

In contrast to conventional pagers, **browse** allows the current file set to
be suspended temporarily while a different set of files is examined, and then
restored at the prior location. This recursive model supports investigation of
logs, search results, source trees, and related materials without discarding
context.

Conceptually, the program may be understood as combining the behavior of
`less` with a stack-based file navigation model similar to `pushd`/`popd`.

## Overview

- Supports examination of multiple files within a single workflow.
- Allows temporary transition to a new file set and later return to the prior
  position.
- Accepts command output from pipelines as browsable input.
- Provides keyboard-driven navigation, search, shell access, and history.
- Preserves context during analysis of logs, source code, and generated output.

## Features

### Navigation

- Forward and reverse paging.
- Continuous scrolling in both directions.
- Horizontal scrolling for wide lines.
- Jump to line numbers.
- Mark pages and return to them.
- Follow and tail modes for changing files.

### Search and Exploration

- Forward and reverse regex and fixed-string search.
- Case-sensitive and case-insensitive search.
- Pattern highlighting.
- Search pattern history.
- Run `grep` on the current file in a nested browse session.

### Multi-File Operation

- Browse multiple files from the command line.
- Browse standard input as a temporary file.
- Open nested file sets with `B`.
- Return from nested file sets with `x` or `X`.
- Rewind the active file list with `Ctrl+R`.
- Show the current remaining file list with `a`.
- Show the suspended browse stack with `A`.

### Additional Facilities

- File and directory completion.
- Shell escape with command completion.
- Persistent file, directory, search, and shell histories.
- Session saving and restoration.
- Transparent browsing of compressed files.
- Run `fmt -s` on the current file in a nested browse session.
- Built-in help screen.

## Continuous Scrolling and Following

**browse** supports continuous scrolling and file following:

- Continuous scrolling moves upward or downward until interrupted. Use `d` to
  scroll toward EOF and `u` to scroll toward SOF.
- When continuous scrolling reaches EOF, **browse** enters follow mode and
  displays lines appended to the file.
- The tail command jumps to EOF and follows subsequent output from that point.
- The cursor indicates whether follow mode is active. In follow mode, the
  cursor is in the lower-left corner; otherwise, it is in the upper-left
  corner.

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
| `-I`, `--fixed-case`  | Search fixed case                             |
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
| `A`      | Print the suspended browse stack                |
| `q`      | Quit current file, save session, continue list  |
| `Q`      | Quit current file without saving, continue list |
| `x`      | Exit current list, save session                 |
| `X`      | Exit current list without saving session        |
| `Ctrl+X` | Exit all nested lists and quit, save session    |
| `Ctrl+Y` | Exit all nested lists and quit, without saving  |

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
file and `~` for the home directory.

When a file set is opened with `B`, **browse** temporarily suspends the current
list. After the nested list is finished, **browse** resumes the previous list
at the prior location.

### Showing the Current List

Press `a` to display the current file and any remaining files in the active
list. For example, if the session began with:

```bash
browse file1 file2 file3
```

and `file2` is currently displayed, pressing `a` shows `file2` and `file3`.

### Showing the Browse Stack

Press `A` to display the current file together with suspended parent files. The
most immediately resumable parent is listed first; for example, while viewing
`file3` after opening `file2` from `file1`, browse shows `[file3] file2 file1`.

### Re-Reading Files

Press `R` to re-read the current file from disk. This is useful when a file is
rewritten in place, replaced, truncated, or otherwise modified in a way not
fully captured by automatic file tracking.

For example:

```bash
mv log log.old
app > log
```

When the current file is moved away or removed, **browse** continues reading
from the already open file descriptor and reports the rescue path in use. If a
new file appears at the original path, press `R` to return to that path and
rebuild the browse state from disk. If the original path is replaced with a
different file, **browse** detects the inode change and automatically reopens
the path. If the file is truncated, **browse** clears the old offsets, prints
`File truncated`, and begins reading the shortened file from the beginning.

### Recovering a Deleted File

Removing a file does not immediately destroy its contents while a process still
holds it open. When the file being browsed is removed, **browse** duplicates
its open descriptor and continues reading, so the file remains readable and can
still be followed.

**browse** reports the descriptor from which the file was rescued:

```text
File removed: recover from /proc/6620/fd/7
```

That path belongs to the **browse** process itself rather than to the invoking
shell, so it remains valid for shell escapes started with `!`. The displayed
filename also becomes the rescue path, which means `%` expands to it and the
contents can be written back to disk:

```bash
!cp % ~/recovered.log
```

Recovery must occur while the file is still open. **browse** closes the rescued
descriptor when you move to another file or leave the session, and the contents
are no longer recoverable after that.

If a new file later appears at the original path, **browse** reports
`File removed: press R to re-read`, leaving the timing of the switch to the
user.

### Rewinding Lists

Press `Ctrl+R` to rewind the active browse list. This returns to the first file
in the current list, including a nested list opened with `B`, without
rewinding any parent list.

### Browsing Compressed Files

**browse** detects compressed files by their magic bytes and decompresses them
transparently. A compressed file may be opened in the same manner as any other
file, either from the command line or with `B` inside an existing session.

Supported formats:

| Format   | Requires     |
| -------- | ------------ |
| gzip     | `gzip`       |
| bzip2    | `bzip2`      |
| xz       | `xz`         |
| zstd     | `zstd`       |
| zip      | `funzip`     |
| lz4      | `lz4`        |
| 7z       | `7z`         |
| compress | `uncompress` |

If the required decompressor is not installed, **browse** displays an error and
skips the file rather than attempting to render raw binary data.

The original compressed filename is recorded in the file history and session
file, so it can be reopened later from the `B` prompt or during a subsequent
launch.

**Limitation:** `Ctrl+R` (rewind list) does not work while browsing a
compressed file. The content is decompressed once into a temporary stream;
rewinding is not supported in that session.

### Changing Directory

Press `C` to change the current working directory. The prompt accepts `~`, `-`,
and `~-`, quoted directory names, and `%` for the parent directory of the
current file. Directory completion includes entries from `CDPATH`.

## Symbol Expansions

Special symbols are expanded in specific prompts; not every symbol is available
in every context:

| Symbol | Expands To                                                |
| ------ | --------------------------------------------------------- |
| `!`    | Last shell command                                        |
| `%`    | Current file name; parent directory of file in `C` prompt |
| `&`    | Current search pattern                                    |
| `~`    | Home directory                                            |
| `-`    | Previous file or previous directory                       |

## Configuration and History

**browse** stores configuration and history in:

```text
~/.browse/
```

The session file is:

```text
~/.browse/browserc
```

Stored session data includes:

- Current file name.
- First line on the page.
- Search pattern.
- Marks.
- Page title.
- Search case-sensitivity mode.
- Fixed-string search mode.

History files are maintained for common workflows, with behavior similar to
Bash history:

- **Shell commands** (`!` key): Every shell command is recorded and can be
  recalled, edited, and rerun.
- **Directory history** (`C` key): Recently visited directories are recorded
  for efficient navigation across large projects.
- **File history** (`B` prompt): Browsed files are saved as full pathnames and
  remain available in current and future sessions.
- **Search patterns** (`/` and `?` prompts): Regex and text search patterns are
  saved so they can be repeated or revisited without retyping them.

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
