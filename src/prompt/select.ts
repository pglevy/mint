/**
 * Interactive select prompt using raw terminal mode.
 *
 * Renders a list of choices with arrow-key navigation.
 * Enter confirms selection, Escape or Ctrl+C cancels (returns null).
 */

const ENCODER = new TextEncoder();
const DECODER = new TextDecoder();

/** ANSI escape sequences for terminal control. */
const ESC = {
  clearLine: "\x1b[2K",
  cursorUp: (n: number) => `\x1b[${n}A`,
  cursorHide: "\x1b[?25l",
  cursorShow: "\x1b[?25h",
  dim: "\x1b[2m",
  reset: "\x1b[0m",
  cyan: "\x1b[36m",
  bold: "\x1b[1m",
};

function write(text: string): void {
  Deno.stderr.writeSync(ENCODER.encode(text));
}

function render(message: string, choices: string[], selected: number): void {
  // Write the prompt message
  write(`${ESC.bold}${message}${ESC.reset}\n`);

  // Write each choice
  for (let i = 0; i < choices.length; i++) {
    if (i === selected) {
      write(`${ESC.cyan}  > ${choices[i]}${ESC.reset}\n`);
    } else {
      write(`${ESC.dim}    ${choices[i]}${ESC.reset}\n`);
    }
  }
}

function clearRendered(lineCount: number): void {
  // Move up and clear each line
  for (let i = 0; i < lineCount; i++) {
    write(ESC.cursorUp(1) + ESC.clearLine);
  }
}

/**
 * Present an interactive selection prompt.
 *
 * @param options.message - The prompt message to display
 * @param options.choices - The list of choices to select from
 * @returns The selected choice string, or null if cancelled
 */
export async function promptSelect(options: {
  message: string;
  choices: string[];
}): Promise<string | null> {
  const { message, choices } = options;

  if (choices.length === 0) {
    return null;
  }

  let selected = 0;
  const totalLines = choices.length + 1; // +1 for the message line

  // Hide cursor and render initial state
  write(ESC.cursorHide);
  render(message, choices, selected);

  try {
    Deno.stdin.setRaw(true);

    const buf = new Uint8Array(64);
    while (true) {
      const n = await Deno.stdin.read(buf);
      if (n === null) {
        // EOF — treat as cancel
        return null;
      }

      const input = DECODER.decode(buf.subarray(0, n));

      // Ctrl+C (0x03) or Escape (0x1b without following bracket sequence of arrow keys)
      if (input === "\x03") {
        return null;
      }

      // Escape key alone (not part of arrow sequence)
      if (input === "\x1b") {
        return null;
      }

      // Enter / Return
      if (input === "\r" || input === "\n") {
        return choices[selected];
      }

      // Arrow key sequences: \x1b[A (up), \x1b[B (down)
      if (input === "\x1b[A") {
        // Up arrow
        selected = selected > 0 ? selected - 1 : choices.length - 1;
        clearRendered(totalLines);
        render(message, choices, selected);
      } else if (input === "\x1b[B") {
        // Down arrow
        selected = selected < choices.length - 1 ? selected + 1 : 0;
        clearRendered(totalLines);
        render(message, choices, selected);
      }
    }
  } finally {
    Deno.stdin.setRaw(false);
    clearRendered(totalLines);
    // Show the selected value as confirmation
    if (choices[selected] !== undefined) {
      write(
        `${ESC.bold}${message}${ESC.reset} ${ESC.cyan}${
          choices[selected]
        }${ESC.reset}\n`,
      );
    }
    write(ESC.cursorShow);
  }
}
