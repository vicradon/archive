<?php

use App\Http\Controllers\AuthController;
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
    Route::resource('/users', UserController::class);
    Route::resource('/projects', ProjectController::class);

    Route::post('/supervisor/register', [UserController::class, 'unboardSupervisor']);
    Route::post('/users/avatar', [UserController::class, 'avatarUpload']);
    Route::get('/projects/status', [ProjectController::class, 'getProjectStatus']);
});

Route::fallback(function () {
    return response()->json([
        'success' => false,
        'message' => 'Route Not Found'
    ], 404);
});
