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

    Usage: browse [-finv] [-p pattern] [-t title] [filename]
     -f, --follow       follow file
     -i, --ignore-case  search ignores case
     -n, --numbers      line numbers
     -p, --pattern      search pattern
     -t, --title        page title
     -v, --version      print version number
     -?, --help         this message

When filename is absent, browse attmpts to restore the session saved in ~/.browserc.

## Scrolling/Following

browse has several scrolling/following modes.

- Scrolling up and down in browse is a continuous process, providing a seamless browsing experience. Once initiated, it persists until you decide to halt it. Consider the scroll and tail commands as toggle switches.

- When scrolling down hits EOF, browse enters follow mode, reading and displaying two lines per read of the input file.

- The tail command jumps to and follows EOF, reading and displaying the input file as fast as browse can read it.

- The cursor position indicates whether or not browse is following the file. If the cursor is in the lower left-hand corner, browse is following. If the cursor is in the upper left-hand corner, browse is idle.

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

## Help Screen

    Command                      Function
    f b [PAGE UP] [PAGE DOWN]    Page down/up
    + - [LEFT] [RIGHT] [ENTER]   Scroll one line
    u d [UP] [DOWN]              Continuous scroll mode
    < >                          Horizontal scroll left/right
    #                            Line numbers
    j                            Jump to line number
    0 ^ [HOME]                   Jump to SOF
    G $ [END]                    Jump to EOF
    z                            Center page on top line
    m                            Mark a page with number 1-9
    1-9                          Jump to mark
    / ?                          Regex search forward/reverse
    n N                          Repeat search forward/reverse
    i                            Case-sensitive search
    &                            Pipe search to grep -nP
    C                            Clear search
    t                            Tail mode
    !                            bash command
    q                            Quit
    Q                            Quit, don't save ~/.browserc    

