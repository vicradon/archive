export function tokenize(expression: string): string[] {
  const tokens: string[] = [];
  const operatorsAndParens = new Set(['+', '-', '*', '/', '(', ')']);
  let i = 0;

  while (i < expression.length) {
    const char = expression[i];

    if (char === ' ') {
      i++;
      continue;
    }

    if (operatorsAndParens.has(char)) {
      tokens.push(char);
      i++;
      continue;
    }

    if (/\d/.test(char)) {
      let num = '';
      while (i < expression.length && /\d/.test(expression[i])) {
        num += expression[i];
        i++;
      }
      tokens.push(num);
      continue;
    }

    throw new Error(`Unexpected character: ${char}`);
  }

  return tokens;
}
