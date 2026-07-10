# Valid Parentheses

**LeetCode:** https://leetcode.com/problems/valid-parentheses/ (LC 20, Easy)

## Description

Given a string containing just `(`, `)`, `{`, `}`, `[`, `]`, determine if the input is valid — every opening bracket must be closed by the same type, in the correct order.

## Idea

Use a stack. Walk the string left to right:

- If you see an opening bracket, push it.
- If you see a closing bracket, check the top of the stack. If it matches the corresponding opener, pop it. If it doesn't match (or the stack is empty), the string is invalid.
- At the end, the string is valid only if the stack is empty (no unclosed openers left).

This works because the *most recently opened* bracket must be the *next one closed* — that's exactly LIFO behavior, which is what a stack gives you for free.

## Complexity

O(n) time, O(n) space (worst case, all openers).

## Edge cases

- Empty string → `true` (vacuously valid)
- Odd length → automatically invalid (good sanity check, can short-circuit early)
- `"([)]"` → invalid (interleaved, not nested)
- `")"` → invalid (pop on empty stack)
