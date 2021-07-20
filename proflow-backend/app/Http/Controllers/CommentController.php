<?php

namespace App\Http\Controllers;

use App\Models\Chapter;
use App\Models\Comment;
use App\Models\Project;
use App\Models\Student;
use App\Notifications\NewComment;
use Illuminate\Http\Request;
use PhpParser\Node\Expr\New_;

class CommentController extends Controller
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
        $validatedData = $request->validate([
            'comment_text' => 'required|string|max:255',
            'page_number' => 'required|numeric',
        ]);
        $supervisor = $request->user()->profile()->first();
        $project = Project::firstWhere('student_id', $request->student_id);
        $chapter = Chapter::find($request->chapter_id);
        $student = Student::find($request->student_id);
        $student_user = $student->user()->first();


        $comment = Comment::create(['comment_text' => $validatedData['comment_text'], 'page_number'  => $validatedData['page_number']]);

        $comment->chapter()->associate($chapter)->save();
        $comment->project()->associate($project)->save();
        $comment->student()->associate($student)->save();
        $comment->supervisor()->associate($supervisor)->save();

        return [$chapter, $validatedData['page_number'], $comment];
        $student_user->notify(new NewComment($chapter, $validatedData['page_number'], $comment));

        return response()->json([
            'success' => true,
            'comment' => $comment
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
        $comment = Comment::find($id);

        if ($request->has('resolve_comment')) {
            $comment->resolved = true;
            $comment->save();

            return response()->json([
                'success' => true,
                'comment' => $comment
            ]);
        }

        $validatedData = $request->validate([
            'comment_text' => 'required|string|max:255'
        ]);

        $comment->comment_text = $validatedData['comment_text'];
        $comment->save();

        return response()->json([
            'success' => true,
            'comment' => $comment
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

    public function getUnresolvedCommentsCount(Request $request)
    {
        $student = $request->user()->profile()->first();
        $comment_count = Comment::where('resolved', false)
            ->where('student_id', $student->id)
            ->count();

        return response()->json([
            'success' => true,
            'comment_count' => $comment_count
        ]);
    }
}
