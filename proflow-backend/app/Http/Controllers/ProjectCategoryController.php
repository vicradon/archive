<?php

namespace App\Http\Controllers;

use App\Models\ProjectCategory;
use Illuminate\Http\Request;

class ProjectCategoryController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index(Request $request)
    {

        $project_categories = $request->has('with_supervisors')
            ? ProjectCategory::leftJoin('supervisors', 'project_categories.id', '=', 'supervisors.project_category_id')->where('supervisors.id', '!=', null)->get()
            : ProjectCategory::all();

        return response()->json([
            'success' => true,
            'project_categories' => $project_categories
        ]);
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
        ]);

        $project_category = ProjectCategory::create(['name' => $validatedData['name']]);

        return response()->json([
            'success' => true,
            'project_category' => $project_category
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
        $project_category = ProjectCategory::find($id);
        return response()->json([
            'success' => true,
            'project_category' => $project_category
        ]);
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
        ]);

        $project_category = ProjectCategory::find($id);
        $project_category->name = $validatedData['name'];
        $project_category->save();

        return response()->json([
            'success' => true,
            'project_category' => $project_category
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
        ProjectCategory::find($id)->delete();
        return response()->json([
            'success' => true,
        ]);
    }
}
