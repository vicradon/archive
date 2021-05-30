<?php

namespace App\Http\Controllers;

use App\Mail\RegisteredSupervisor;
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
        $users = User::all();

        // return $request->user()->can('viewAny', $users);

        // if ($request->user()->cannot('view', $users)) {
        //     return response(['success' => false, 'message' => "You cannot view other user's data"], 403);
        // }

        return ['success' => true, 'data' => $users];
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
            'supervisor_category' => 'required|string',
            'department' => 'required|string'
        ]);

        $password = Str::uuid();

        $user = User::create([
            'name' => $validatedData['first_name'] . " " . $validatedData['last_name'],
            'email' => $validatedData['email'],
            'password' => Hash::make($password),
            'department' => $validatedData['department']
        ]);

        Mail::to($user)->send(new RegisteredSupervisor($user, $request->user(), $password));
        return response()->json([
            'success' => true,
            'message' => "An email with further instructions has been sent to the supervisor's email"
        ], 201);
    }
    public function supervisorPairing($projectCategoryId)
    {
        // $supervisor_with_min_assigned_students = DB::table('users')->select(DB::raw('select * from users where count()'))
        // $lowest_number_of_assigned_students = 0;
        // $available_supervisor = null;

        /*
        user_id => number of students with this id as their supervisor id
        3 => 4
        7 => 2
        14 => 1

        */

        // foreach ($target_supervisors as $target_supervisor){
        //     if $target_supervisor->
        // }
    }
}
