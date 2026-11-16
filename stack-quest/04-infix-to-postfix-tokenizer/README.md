# Infix → Postfix → Evaluate (Shunting-Yard + Tokenizer)

This isn't a single official LeetCode problem — it's the "other half" of RPN that we built as a
follow-up exercise. It's the classic **Shunting-Yard algorithm** (invented by Dijkstra), and it's
closely related to these LeetCode problems:

- **Basic Calculator** — https://leetcode.com/problems/basic-calculator/ (LC 224, Hard)
- **Basic Calculator II** — https://leetcode.com/problems/basic-calculator-ii/ (LC 227, Medium)

## Description

Convert infix notation (what you normally write, e.g. `2 + 1 * 3`) into postfix (`2 1 3 * +`),
then evaluate the postfix result — building a full raw-string → tokens → postfix → answer pipeline.

## The core idea (Shunting-Yard)

Operators need to wait until we know both their operands are ready — and precedence/parentheses
determine *when* an operator gets to "fire." A stack holds operators that are waiting to be
output, and you pop them out (into the postfix result) whenever a lower- or equal-precedence
operator arrives, or when you hit a closing parenthesis.

Algorithm, reading left to right through the infix tokens:

1. If token is a **number** → append it directly to output.
2. If token is an **operator** → pop from the operator stack into output while the top of the
   stack has greater-or-equal precedence, then push the current operator.
3. If token is `(` → push it onto the operator stack.
4. If token is `)` → pop operators into output until you hit the matching `(`, then discard the `(`.
5. At the end, pop any remaining operators into output.

## Pipeline

```
raw string  --tokenize()-->  tokens  --infixToPostfix()-->  postfix tokens  --evalRPN()-->  number
```

`evaluateInfix()` chains all three together.

## Files

- `tokenize.ts` — splits a raw infix string into tokens (numbers, operators, parens), keeping
  multi-digit numbers together.
- `infixToPostfix.ts` — Shunting-Yard conversion from infix tokens to postfix tokens.
- `evalRPN.ts` — evaluates postfix tokens (same logic as Evaluate Reverse Polish Notation).
- `evaluateInfix.ts` — combines all three into a single `evaluateInfix(expression: string): number`.

## What the tokenizer deliberately doesn't handle (yet)

- Unary minus (e.g. `-5 + 3`)
- Decimal numbers
- Whitespace-insensitive multi-char operators (`**`, `//`, etc.)

## Complexity

O(n) time and space for tokenizing, converting, and evaluating.
