<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

class CreateSupervisorsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('supervisors', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('coordinator_id')->nullable();
            $table->foreign('coordinator_id')->references('id')->on('coordinators')->onDelete('CASCADE');
            $table->unsignedBigInteger('project_category_id')->nullable();
            $table->foreign('project_category_id')->references('id')->on('project_categories')->onDelete('CASCADE');
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('supervisors');
    }
}
