<?php

use App\Http\Controllers\AuthController;
use App\Http\Controllers\ChapterController;
use App\Http\Controllers\CommentController;
use App\Http\Controllers\ProjectCategoryController;
use App\Http\Controllers\UserController;
use App\Http\Controllers\ProjectController;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\File;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| is assigned the "api" middleware group. Enjoy building your API!
|
*/

Route::middleware('auth:api')->get('/user', function (Request $request) {
    return $request->user();
});

Route::get('/docs',  function () {
    return File::get("openapi.json");
});
Route::post('/register', [AuthController::class, 'register']);
Route::post('/login', [AuthController::class, 'login']);

Route::middleware(['auth:sanctum'])->group(function () {
    Route::apiResource('/users', UserController::class);
    Route::apiResource('/projects', ProjectController::class);
    Route::apiResource('/comments', CommentController::class);
    Route::apiResource('/chapters', ChapterController::class);
    Route::apiResource('/project-categories', ProjectCategoryController::class);

    Route::get('/project', [ProjectController::class, 'getProject']);
    Route::get('/recent-proposals', [ProjectController::class, 'recentProposals']);
    Route::get('/recent-proposals/{proposal_id}', [ProjectController::class, 'recentProposal']);
    Route::patch('/recent-proposals/{proposal_id}', [ProjectController::class, 'updateProposal']);
    Route::post('/projects/upload', [ProjectController::class, 'uploadProject']);

    Route::get('/supervisors', [UserController::class, 'getSupervisors']);
    Route::patch('/supervisor/password', [UserController::class, 'updateSupervisorPassword']);
    Route::post('/supervisor/register', [UserController::class, 'unboardSupervisor']);
    Route::get('/supervisor/students', [UserController::class, 'studentsUnderSupervisor']);
    Route::post('/users/avatar', [UserController::class, 'avatarUpload']);
    Route::get('/project-status', [ProjectController::class, 'getProjectStatus']);
    Route::get('/comments-count', [CommentController::class, 'getUnresolvedCommentsCount']);

    // remove this during refactoring: use /chapters endpoint
    Route::get('/students/{student_id}/chapters', [ProjectController::class, 'getStudentChapters']);
    // remove this during refactoring: use /chapters/{chapter_id} endpoint
    Route::get('/students/{student_id}/chapters/{chapter_id}', [ProjectController::class, 'getStudentChapter']);
});

Route::get('/stuff', function () {
    $frontend_root = config('app.env') === 'local' ?
        config('app.frontend_local_root_url') :
        config('app.frontend_remote_root_url');
    return $frontend_root;
});

Route::fallback(function () {
    return response()->json([
        'success' => false,
        'message' => 'Route Not Found'
    ], 404);
});
