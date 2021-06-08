<?php

namespace App\Http\Controllers;

use App\Http\Resources\UserCollection;
use App\Http\Resources\UserResource;
use App\Mail\RegisteredSupervisor;
use App\Models\ProjectCategory;
use App\Models\User;
use Exception;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Str;
use Illuminate\Auth\Events\Registered;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Mail;

// use Symfony\Component\HttpFoundation\Response;

class UserController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index(Request $request)
    {

        $request->validate([
            'role' => 'string|in:student,supervisor',
            'name' => 'string|max:255',
            'email' => 'string|email|max:255',
            'supervisor_category' => 'in:iot,transportation,communication',
            'student_under_supervisor' => 'numeric'
        ]);

        $users = User::where('email', '!=', NULL);

        /* supervisors
            - by category
            - by name
            - by email
        */

        /* students of a supervisor
            - By accepted/rejected projects
            - By name
        */

        if ($request->has('role')) {
            if ($request->role === 'supervisor') {
                $users->where('profile_type', 'App\Models\Supervisor');
            } else if ($request->role === 'student') {
                $users->where('profile_type', 'App\Models\Student');
            }
        }
        if ($request->has('name')) {
            $users->where('name', 'like', "%{$request->name}%");
        }
        if ($request->has('email')) {
            $users->where('email', $request->email);
        }

        $users = $users->get();

        # supervisors with same project category
        if ($request->has('supervisor_category')) {
            global $project_category_id;
            $project_category_id = ProjectCategory::firstWhere('name', $request->supervisor_category)->id;

            $users = $users->filter(function ($user) {
                $profile = $user->profile()->first();
                global $project_category_id;
                return $profile->project_category_id === $project_category_id;
            });
        }

        if ($request->has('students_under_supervisor')) {
            global $supervisor_id;
            $supervisor_id = (int)($request->students_under_supervisor);

            $users = $users->filter(function ($user) {
                $profile = $user->profile()->first();
                global $supervisor_id;
                return $profile->supervisor_id === $supervisor_id;
            });
        }

        return UserResource::collection($users);
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        $validatedData = $request->validate([
            'name' => 'required|string|max:255',
            'email' => 'required|string|email|max:255|unique:users',
            'password' => 'required|string|min:8',
        ]);

        $user = User::create([
            'name' => $validatedData['name'],
            'email' => $validatedData['email'],
            'password' => Hash::make($validatedData['password']),
        ]);

        return $user;
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show($id)
    {
        $user = User::find($id);
        return $user;
    }

    /**
     * Update the specified resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function update(Request $request, $id)
    {
        $validatedData = $request->validate([
            'name' => 'required|string|max:255',
            'email' => 'required|string|email|max:255|unique:users',
            'password' => 'required|string|min:8',
        ]);

        $user = User::find($id);
        $user->update([
            'name' => $validatedData['name'],
            'email' => $validatedData['email'],
            'password' => Hash::make($validatedData['password']),
        ]);
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        User::destroy($id);
    }

    public function avatarUpload(Request $request)
    {
        $user = $request->user();
        $avatar_path = $request->file('avatar')->store('avatars');
        $user->avatar_url = $avatar_path;
        $user->save();

        $avatar_url = $_SERVER['HTTP_HOST'] . "/" . $avatar_path;
        return response()->json(["avatar_url" => $avatar_url]);
    }


    public function unboardSupervisor(Request $request)
    {
        $authzResponse = Gate::inspect('unboardSupervisor', $request->user());

        if ($authzResponse->denied()) {
            return response()->json([
                'success' => false,
                'message' => $authzResponse->message()
            ], 403);
        }

        $validatedData = $request->validate([
            'first_name' => 'required|string|max:255',
            'last_name' => 'required|string|max:255',
            'email' => 'required|string|email|max:255|unique:users',
            'project_category' => 'required|string',
            'department' => 'required|string'
        ]);

        $password = Str::uuid();

        $supervisor = User::create([
            'name' => $validatedData['first_name'] . " " . $validatedData['last_name'],
            'email' => $validatedData['email'],
            'password' => Hash::make($password),
            'department' => $validatedData['department']
        ]);

        $projectCategory = ProjectCategory::findWhere('name', $validatedData['project_category']);
        $supervisor->projectCategory()->associate($projectCategory)->save();


        Mail::to($supervisor)->send(new RegisteredSupervisor($supervisor, $request->user(), $password));
        return response()->json([
            'success' => true,
            'message' => "An email with further instructions has been sent to the supervisor's email"
        ], 201);
    }
}

// Student::has('supervisor')->with('user')->get();
// User::has('profile')->with('profile')->get();