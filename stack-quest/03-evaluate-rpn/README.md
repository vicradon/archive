# Evaluate Reverse Polish Notation

**LeetCode:** https://leetcode.com/problems/evaluate-reverse-polish-notation/ (LC 150, Medium)

## Description

Evaluate an arithmetic expression given in Reverse Polish (postfix) notation. Tokens are numbers or operators (`+`, `-`, `*`, `/`), e.g. `["2","1","+","3","*"]` evaluates to `9` (i.e. `(2+1)*3`).

## Idea

In postfix notation, an operator always applies to the two numbers immediately before it. So: walk the tokens left to right, push numbers onto a stack, and whenever you hit an operator, pop the top two numbers, apply the operator, and push the result back. At the end, the stack has exactly one value — the answer.

The order of the pop matters: the *first* pop is the right-hand operand, the *second* pop is the left-hand operand (since the top of the stack is the most recently pushed number).

## Complexity

O(n) time, O(n) space.
