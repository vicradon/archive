<?php

namespace Database\Seeders;

use App\Models\Coordinator;
use App\Models\ProjectCategory;
use App\Models\Student;
use App\Models\Supervisor;
use Illuminate\Database\Seeder;

class UserSeeder extends Seeder
{
    /**
     * Run the database seeds.
     *
     * @return void
     */
    public function run()
    {

        $project_categories = ProjectCategory::all();
        Student::factory()->count(20)->hasUser()->create();
        for ($i = 0; $i < 5; $i++) {
            $project_category_index = rand(0, 2);
            Supervisor::factory()->for($project_categories[$project_category_index])->hasUser()->create();
        }
        Coordinator::factory()->count(2)->hasUser()->create();
    }
}
