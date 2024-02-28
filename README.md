# browse

A simple, unconventional file browser.

## Goals

- Create a file browser with only the most common functions
- Keep it simple, keep it friendly

## Features

- Forward and reverse paging
- Forward and reverse scrolling
- Continuous forward and reverse scrolling
- Horizontal scrolling
- Jump to lines
- Mark pages
- Forward and reverse searches by regular expression
- tail -f
- Line numbers
- Save session
- Help screen

## Usage

```text
    Usage: browse [-fn] [-p pattern] [-t title] [filename]
     -f, --follow   follow file
     -n, --numbers  line numbers
     -p, --pattern  search pattern
     -t, --title    page title
     -v, --version  print version number
     -?, --help     this message
```

When filename is absent, browse attmpts to restore the session saved in ~/.browserc.

## Scrolling/Following

browse has several scrolling/following modes.

- Scrolling up and down is continuous, meaning once started, scrolling continues until
it is instructed to stop.  Think of the scroll and tail commands as toggle switches.

- When scrolling down hits EOF, browse enters follow mode, reading and displaying
the input file two lines at a time.

- The tail command jumps to and follows EOF, reading and displaying the
input file up to 256 lines at a time.

- The cursor position indicates whether or not browse is following the file.  If the
cursor is in the lower left-hand corner, browse is following.  If the cursor is in
the upper left-hand corner, browse is idle.

## Saved Sessions

browse saves sessions in ~/.browserc.  The format of the file is plaintext containing the following lines:

 1. file name
 2. first row on page
 3. search pattern
 4. marks
 5. page title

The session attributes not saved:

- search direction
- numbers
- bash command
- horizontal scroll
- follow/tail mode

browse does not save sessions when the input is standard in, or when browse exits with the Q command.

## Limitations

- Xterm specific
- Logical lines chopped to the screen width
- Probably US-centric
- Can be confused by lines with non-printable characters
