import { tokenize } from './tokenize';
import { infixToPostfix } from './infixToPostfix';
import { evalRPN } from './evalRPN';

export function evaluateInfix(expression: string): number {
  const tokens = tokenize(expression);
  const postfix = infixToPostfix(tokens);
  return evalRPN(postfix);
}

// Example:
const result = evaluateInfix('(2 + 1) * 3') // -> 9
console.log(result)