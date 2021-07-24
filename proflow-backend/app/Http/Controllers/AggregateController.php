<?php

namespace App\Http\Controllers;

use App\Models\Project;
use App\Models\ProjectCategory;
use App\Models\Student;
use App\Models\Supervisor;
use Illuminate\Http\Request;

class AggregateController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        //
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        //
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
        //
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

    public function getSummary(Request $request)
    {
        $supervisor_count = Supervisor::count();
        $student_count = Student::count();
        $project_count = Project::count();
        $project_category_count = ProjectCategory::count();

        return response()->json([
            'success' => true,
            'summary' => [
                'supervisor_count' => $supervisor_count,
                'student_count' => $student_count,
                'project_count' => $project_count,
                'project_category_count' => $project_category_count
            ]
        ]);
    }
}
