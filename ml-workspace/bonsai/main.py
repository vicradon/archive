from mlx_lm import load, generate

model, tokenizer = load("prism-ml/Bonsai-4B-mlx-1bit")

response = generate(
    model,
    tokenizer,
    prompt="Explain quantum computing in simple terms.",
    max_tokens=256,
)
print(response)

