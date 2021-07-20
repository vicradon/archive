<?php

namespace App\Http\Controllers;

use App\Http\Resources\ProjectResource;
use App\Models\Chapter;
use App\Models\Comment;
use App\Models\Project;
use App\Models\ProjectCategory;
use App\Models\Student;
use App\Models\Supervisor;
use App\Models\User;
use App\Notifications\NewProjectUpload;
use App\Notifications\ProposalApproved;
use App\Notifications\ProposalRejected;
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

        // if ($project_exists) {
        //     return response()->json([
        //         'success' => false,
        //         'message' => "You already have an active project"
        //     ], 403);
        // }

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

        $supervisor_to_be_paired = $this->supervisorPairing($project);

        if (!$supervisor_to_be_paired) {
            return response()->json([
                'success' => true,
                'message' => 'No supervisor has been assigned to your project category',
                'supervisor_name' => null
            ], 201);
        }

        $supervisor = Supervisor::find($supervisor_to_be_paired->id);
        $supervisor->student()->save($student);
        $project->supervisor()->associate($supervisor)->save();
        $supervisor_user_profile = $supervisor->user()->first();

        try {
            $supervisor_user_profile->notify(
                new ProjectProposed(
                    $supervisor_user_profile,
                    $request->user(),
                    $project
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

    public function supervisorPairing($project)
    {
        # Look for the supervisor under the project_category_id with the least number of students

        # Select the first available supervisor with no assigned students
        // $available_supervisor_query = DB::table('supervisors')
        //     ->select('supervisors.id', 'supervisors.project_category_id')
        //     ->leftJoin('students', 'supervisors.id', '=', 'students.supervisor_id')
        //     // ->where('students.id', NULL)
        //     ->where('supervisors.project_category_id', $project->project_category_id)
        //     ->first();

        $available_supervisor = Supervisor::where('project_category_id', $project->project_category_id)
            ->leftJoin('students', 'supervisors.id', '=', 'students.supervisor_id')
            ->withCount('student')
            ->orderBy('student_count', 'asc')
            ->first();
        if (!$available_supervisor) {
            return null;
        }

        return $available_supervisor;
    }

    public function getProjectStatus(Request $request)
    {
        $user_data  = $request->user();
        $student = $user_data->profile()->first();
        $supervisor = $student->supervisor_id ? User::where('profile_id', $student->supervisor_id)->where('profile_type', "App\Models\Supervisor")->first() : null;

        $project = Project::firstWhere('student_id', $student->id);

        return response()->json([
            'success' => true,
            'status' => $project ? $project->status : null,
            'supervisor_name' => $supervisor ? $supervisor->name : null,
            'avatar_url' => $request->has('with_avatar_url') ? $user_data->avatar_url : null
        ]);
    }

    function isLocalhost($whitelist = ['127.0.0.1', '::1'])
    {
        return in_array($_SERVER['REMOTE_ADDR'], $whitelist);
    }

    public function uploadProject(Request $request)
    {
        $page_ranges = json_decode($request->page_ranges, true);
        $student = $request->user()->profile()->first();
        $supervisor_user = $student->supervisor()->first()->user()->first();
        $project = $student->project()->first();
        $file_path = $request->file('project_pdf')->store('public/project_pdfs');

        $pdf_path = config('app.url') . "/storage" . substr($file_path, 6);

        $supervisor_user->notify(new NewProjectUpload($request->user()));

        $old_path = $project->pdf_path;

        if ($old_path) {
            $path_without_storage = explode("/storage", $old_path)[1];
            Storage::delete("/public" . $path_without_storage);
        }

        # Delete old chapters and create new chapters
        Chapter::where('student_id', $student->id)->delete();

        for ($i = 1; $i <= $request->available_chapters; $i++) {
            $start_page = $page_ranges[$i][0];
            $end_page = $page_ranges[$i][1];

            $chapter = Chapter::create(['serial_number' => $i, 'start_page' => $start_page, 'end_page' => $end_page]);
            $student->chapters()->save($chapter);
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
        $student_user = Student::where('supervisor_id', $proposal->supervisor_id)->first()->user()->first();

        if (!$proposal) {
            return response()->json([
                'success' => false,
                'message' => "The proposal with this ID does not exist or has already been approved"
            ], 404);
        }

        if ($validatedData['action'] === 'approve') {
            $proposal->status = "approved";
            $student_user->notify(new ProposalApproved($proposal));
        } else if ($validatedData['action'] === 'reject') {
            $proposal->status = "rejected";
            $student_user->notify(new ProposalRejected($proposal));
        }

        $proposal->save();


        return response()->json([
            'success' => true,
            'proposal' => $proposal
        ]);
    }
}
