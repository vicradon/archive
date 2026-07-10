const PRECEDENCE: Record<string, number> = {
  '+': 1,
  '-': 1,
  '*': 2,
  '/': 2,
};

export function infixToPostfix(tokens: string[]): string[] {
  const output: string[] = [];
  const opStack: string[] = [];

  for (const token of tokens) {
    if (token === '(') {
      opStack.push(token);
    } else if (token === ')') {
      while (opStack.length > 0 && opStack[opStack.length - 1] !== '(') {
        output.push(opStack.pop()!);
      }
      opStack.pop(); // discard the '('
    } else if (token in PRECEDENCE) {
      while (
        opStack.length > 0 &&
        opStack[opStack.length - 1] !== '(' &&
        PRECEDENCE[opStack[opStack.length - 1]] >= PRECEDENCE[token]
      ) {
        output.push(opStack.pop()!);
      }
      opStack.push(token);
    } else {
      // number
      output.push(token);
    }
  }

  while (opStack.length > 0) {
    output.push(opStack.pop()!);
  }

  return output;
}
