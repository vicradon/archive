export function evalRPN(tokens: string[]): number {
  const stack: number[] = [];
  const operators = new Set(['+', '-', '*', '/']);

  for (const token of tokens) {
    if (operators.has(token)) {
      const right = stack.pop()!;
      const left = stack.pop()!;
      let result: number;

      switch (token) {
        case '+':
          result = left + right;
          break;
        case '-':
          result = left - right;
          break;
        case '*':
          result = left * right;
          break;
        case '/':
          result = Math.trunc(left / right);
          break;
        default:
          throw new Error(`Unknown operator: ${token}`);
      }

      stack.push(result);
    } else {
      stack.push(Number(token));
    }
  }

  return stack.pop()!;
}
