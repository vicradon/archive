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
use Illuminate\Support\Facades\Storage;

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

        $project = Project::create(['name' => $validatedData['name'], 'description' => $validatedData['description']]);
        $project_category = ProjectCategory::where('name', $validatedData['category'])->first();
        $student = $request->user()->profile()->first();

        $project->projectCategory()->associate($project_category)->save();
        $project->student()->associate($student)->save();

        $supervisor = $this->supervisorPairing($student, $project);

        if (!$supervisor) {
            return response()->json([
                'success' => true,
                'message' => 'No supervisor has been assigned to your project category',
                'supervisor_name' => null
            ], 201);
        }

        $project->supervisor()->associate($supervisor)->save();
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
                'message' => 'supervisor email is not valid so no email was sent',
                'supervisor_name' => $supervisor_user_profile->name
            ], 201);
        }

        return response()->json([
            'success' => true,
            'supervisor_name' => $supervisor_user_profile->name
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
            'description' => 'required|string|max:1024'
        ]);

        # Update the name and description from the validated data
        $project = Project::find($id);

        $project->name = $validatedData['name'];
        $project->description = $validatedData['description'];
        $project->status = "pending";

        $request->user()->profile()->first()->project()->save($project);
        return response()->json([
            'success' => true
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
            return null;
        }

        $available_supervisor = Supervisor::firstWhere('id', $available_supervisor_query->id);
        $available_supervisor->student()->save($student);
        return $available_supervisor;
    }

    public function getProjectStatus(Request $request)
    {
        $student = $request->user()->profile()->first();
        $supervisor = $student->supervisor_id ? User::where('profile_id', $student->supervisor_id)->where('profile_type', "App\Models\Supervisor")->first() : null;


        $project = Project::firstWhere('student_id', $student->id);

        return response()->json([
            'success' => true,
            'status' => $project ? $project->status : null,
            'supervisor_name' => $supervisor ? $supervisor->name : null
        ]);
    }

    public function uploadProject(Request $request)
    {
        $project = $request->user()->profile()->first()->project()->first();
        $pdf_path = $request->file('project_pdf')->store('public/project_pdfs');
        $old_file = $project->pdf_path;

        if ($old_file) {
            Storage::delete($old_file);
        }
        $project->pdf_path = $pdf_path;
        $project->save();

        return response()->json([
            'success' => true,
        ]);
    }

    public function getProject(Request $request)
    {
        $project = $request->user()->profile()->first()->project()->with('projectCategory')->first();

        return response()->json([
            'success' => true,
            'project' => $project
        ]);
    }

    public function recentProposals(Request $request)
    {

        $supervisor_profile = $request->user()->profile()->first();
        # Return the project, the student that has the project along with their user data
        $proposals = Project::where('status', 'pending')->where('supervisor_id', $supervisor_profile->id)->with('student', 'student.user')->paginate(10);

        return $proposals;
    }

    public function recentProposal(Request $request, $proposal_id)
    {
        $proposal = Project::Where('id', $proposal_id)->with('student', 'student.user')->with('projectCategory')->first();

        if (!$proposal) {
            return response()->json([
                'success' => false,
                'message' => "The proposal with this ID does not exist or has already been approved"
            ], 404);
        }

        return response()->json([
            'success' => true,
            'proposal' => $proposal
        ]);
    }

    public function updateProposal(Request $request, $proposal_id)
    {
        $validatedData = $request->validate([
            'action' => 'required|string|in:approve,reject'
        ]);

        $proposal = Project::firstWhere('id', $proposal_id);

        if (!$proposal) {
            return response()->json([
                'success' => false,
                'message' => "The proposal with this ID does not exist or has already been approved"
            ], 404);
        }

        if ($validatedData['action'] === 'approve') {
            $proposal->status = "approved";
        } else if ($validatedData['action'] === 'reject') {
            $proposal->status = "rejected";
        }

        $proposal->save();

        return response()->json([
            'success' => true,
            'proposal' => $proposal
        ]);
    }
}

# Se