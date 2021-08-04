<?php

namespace App\Http\Controllers;

use App\Models\Chapter;
use App\Models\Comment;
use App\Models\Project;
use Illuminate\Http\Request;

class ChapterController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index(Request $request)
    {
        $student_id = $request->has('student_id') ? $request->student_id : $request->user()->profile()->first()->id;
        $chapters = Chapter::where('chapters.student_id', $student_id)
            ->leftJoin('comments', 'chapters.id', '=', 'comments.chapter_id')
            ->withCount([
                'comments',
                'comments as comments_count' => function ($query) {
                    $query->where('resolved', false);
                }
            ])
            ->distinct()
            ->get();

        return response()->json([
            'success' => true,
            'chapters' => $chapters
        ]);
    }

    public function getStudentChapters(Request $request)
    {
        $chapters = Chapter::where('chapters.student_id', $request->student_id)
            ->leftJoin('comments', 'chapters.id', '=', 'comments.chapter_id')
            ->withCount([
                'comments',
                'comments as comments_count' => function ($query) {
                    $query->where('resolved', false);
                }
            ])
            ->distinct()
            ->get();

        return response()->json([
            'success' => true,
            'chapters' => $chapters
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
        //
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show(Request $request, $id)
    {
        $student_id = $request->has('student_id') ? $request->student_id : $request->user()->profile()->first()->id;
        $chapter = Chapter::find($id);
        $project = Project::where('student_id', $student_id)->first();
        $comments = Comment::where('student_id', $student_id)->where('resolved', false)->where('chapter_id', $id)->get();

        return response()->json([
            'success' => true,
            'chapter' => $chapter,
            'pdf_path'  => $project->pdf_path,
            'comments' => $comments,
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
            'status' => 'required|string|in:pending,approve'
        ]);

        $chapter = Chapter::find($id);
        $chapter->status = "approved";

        $chapter->save();
        return response()->json([
            'success' => true,
            'chapter' => $chapter
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
}
