<?php

namespace App\Http\Controllers;

use App\Http\Resources\ProjectResource;
use App\Models\Project;
use App\Models\ProjectCategory;
use App\Models\Student;
use App\Models\Supervisor;
use App\Models\User;
use App\Notifications\ProjectProposed;
use Exception;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class ProjectController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        $student = $request->user()->profile()->first();
        $project_exists = Project::firstWhere('student_id', $student->id);

        if ($project_exists) {
            return response()->json([
                'success' => false,
                'message' => "You already have an active project"
            ], 403);
        }

        $validatedData = $request->validate([
            'name' => 'required|string|max:255',
            'category' => 'required|string|max:255',
            'description' => 'required|string|max:1024'
        ]);

        $project_category = ProjectCategory::where('name', $validatedData['category'])->first();

        $project = Project::create(['name' => $validatedData['name'], 'description' => $validatedData['description']]);
        $project_category->project()->save($project);
        $student = $request->user()->profile()->first();
        $student->project()->save($project);
        $supervisor = $this->supervisorPairing($student, $project);

        if (!$supervisor) {
            return response()->json([
                'success' => true,
                'message' => 'No supervisor has been assigned to your project category',
                'supervisor_id' => null
            ], 201);
        }

        $supervisor_user_profile = $supervisor->user()->first();

        try {
            $supervisor_user_profile->notify(
                new ProjectProposed(
                    $supervisor_user_profile,
                    $request->user(),
                    $project,
                    ['new_proposal' => true]
                )
            );
        } catch (Exception $e) {
            return response()->json([
                'success' => true,
                'message' => 'email not valid so was not sent',
                'supervisor_id' => $supervisor_user_profile->id
            ], 201);
        }

        return response()->json([
            'success' => true,
            'supervisor_id' => $supervisor_user_profile->id
        ], 201);
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show($id)
    {
        //
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
        # If category changed, delete the supervisor id from the requesting student and find a new supervisor
        # Else, return the same supervisor

        $validatedData = $request->validate([
            'name' => 'required|string|max:255',
            'category' => 'required|string|max:255',
            'description' => 'required|string|max:1024'
        ]);
        # Update the name and description from the validated data
        $project = Project::find($id);
        $project->name = $validatedData['name'];
        $project->description = $validatedData['description'];

        $student = $request->user()->profile()->project()->first();
        $student->save($project);

        # update the project category on the project if category was changed
        $project_category = ProjectCategory::where('name', $validatedData['category'])->first();
        $project_category->project()->save($project);


        $supervisor = $request->user()->profile()->supervisor()->first();


        if ($project_category->id !== $project->category_id) {
            # unlink the old supervisor from the student
            $student->supervisor()->dissociate()->save();
        }
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        //
    }

    public function supervisorPairing($student, $project)
    {
        # Look for the supervisor under the project_category_id with the least number of students

        # Select the first available supervisor with no assigned students
        $available_supervisor_query = DB::table('supervisors')
            ->select('supervisors.id', 'supervisors.project_category_id')
            ->leftJoin('students', 'supervisors.id', '=', 'students.supervisor_id')
            ->where('students.id', NULL)
            ->where('supervisors.project_category_id', $project->project_category_id)
            ->first();

        if (!$available_supervisor_query) {
            return $available_supervisor_query;
        }

        $available_supervisor = Supervisor::firstWhere('id', $available_supervisor_query->id);
        $available_supervisor->student()->save($student);
        return $available_supervisor;

        # select first supervisor with least number of students
        // $target_supervisor_id = DB::table('students')
        //     ->select(DB::raw('count(supervisor_id) as supervisor_id_count, supervisor_id'))
        //     ->where('supervisor_id', '!=', NULL)
        //     ->where('supervisors.project_category_id', $project->project_category_id)
        //     ->groupBy('supervisor_id')
        //     ->orderBy('supervisor_id_count')
        //     ->first()->supervisor_id;


        // $target_supervisor = Supervisor::firstWhere('id', $target_supervisor_id);
        // $target_supervisor->student()->save($student);
        // return $target_supervisor;
    }

    public function getProjectStatus(Request $request)
    {
        $student_id = $request->user()->profile()->first()->id;
        Log::info($student_id);
        $status = Project::firstWhere('student_id', $student_id)->status;

        return response()->json([
            'success' => true,
            'status' => $status,
        ]);
    }
}
