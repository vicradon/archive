<?php

namespace App\Http\Controllers;

use App\Models\Coordinator;
use App\Models\Role;
use App\Models\Student;
use Illuminate\Support\Facades\Hash;
use Illuminate\Http\Request;
use App\Models\User;
use Illuminate\Support\Facades\Auth;
use Illuminate\Auth\Events\Registered;
use Illuminate\Support\Facades\Log;

class AuthController extends Controller
{
    public function register(Request $request)
    {
        $validatedData = $request->validate([
            'first_name' => 'required|string|max:255',
            'last_name' => 'required|string|max:255',
            'email' => 'required|string|email|max:255|unique:users',
            'password' => 'required|string|min:8',
            'reg_number' => 'exclude_if:role,coordinator|required|numeric|unique:students',
            'role' => 'required|in:student,coordinator',
            'department' => 'required|string|max:255',
            'avatar' => 'mimes:jpg,jpeg,png|max:2048',
        ]);

        $user = User::create([
            'name' => $validatedData['first_name']  . " " . $validatedData['last_name'],
            'email' => $validatedData['email'],
            'password' => Hash::make($validatedData['password']),
            'department' => $validatedData['department'],
        ]);

        if (isset($validatedData['avatar'])) {
            $avatar_path = $request->file('avatar')->storeAs('storage/avatars', $request->file('avatar')->getClientOriginalName());
            $user->avatar_url = $avatar_path;
        }

        if ($validatedData['role'] === 'student') {
            $student = Student::create(['reg_number' => $validatedData['reg_number']]);
            $student->user()->save($user);
        } else if ($validatedData['role'] === 'coordinator') {
            $coordinator = Coordinator::create();
            $coordinator->user()->save($user);
        }

        $token = $user->createToken('auth_token')->plainTextToken;

        $user->avatar_url = $_SERVER['HTTP_HOST'] . "/" . $user->avatar_url;
        event(new Registered($user));

        return response()->json([
            'access_token' => $token,
            'token_type' => 'Bearer',
            'user' => $user,
            'success' => true
        ], 201);
    }

    public function login(Request $request)
    {
        $validatedData = $request->validate([
            'email' => 'required|string|email|max:255',
            'password' => 'required|string|min:8',
        ]);

        if (!Auth::attempt(['email' => $validatedData['email'], 'password' => $validatedData['password']])) {
            return response()->json([
                'message' => 'Invalid login details'
            ], 401);
        }

        $user = User::where('email', $request['email'])->firstOrFail();

        $token = $user->createToken('auth_token')->plainTextToken;

        $user->avatar_url = $_SERVER['HTTP_HOST'] . "/" . $user->avatar_url;
        return response()->json([
            'access_token' => $token,
            'token_type' => 'Bearer',
            'user' => $user,
            'success' => true
        ]);
    }
    public function logout(Request $request)
    {
        $request->user()->currentAccessToken()->delete();
        return response()->json([
            'success' => true,
            'message' => 'logged out successfully',
        ]);
    }
}
