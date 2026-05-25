import numpy as np

# precompute the schedule once
def get_alpha_bar(T):
    betas = np.linspace(1e-4, 0.02, T)  # common linear beta schedule
    alphas = 1 - betas
    alpha_bar = np.cumprod(alphas)       # running product
    return alpha_bar

alpha_bar = get_alpha_bar(T=20)
root_alpha_bar = np.sqrt(alpha_bar)
root_1_minus_alpha_bar = np.sqrt(1 - alpha_bar)

print(alpha_bar)
print(root_alpha_bar)
print(root_1_minus_alpha_bar)
