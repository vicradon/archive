function dailyTemperatures(temperatures: number[]): number[] {
  const answer: number[] = new Array(temperatures.length).fill(0);
  const stack: number[] = []; // indices, decreasing temperature bottom to top

  for (let i = 0; i < temperatures.length; i++) {
    while (
      stack.length > 0 &&
      temperatures[i] > temperatures[stack[stack.length - 1]]
    ) {
      const poppedIndex = stack.pop()!;
      answer[poppedIndex] = i - poppedIndex;
    }
    stack.push(i);
  }

  return answer;
}

console.log(dailyTemperatures([73,74,75,71,69,72,76,73])) // expected [1,1,4,2,1,1,0,0]