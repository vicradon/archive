<?php

namespace App\Mail;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Mail\Mailable;
use Illuminate\Queue\SerializesModels;

class RegisteredSupervisor extends Mailable
{
    use Queueable, SerializesModels;

    /**
     * Create a new message instance.
     *
     * @return void
     */
    public function __construct($user, $coordinator, $password)
    {
        $this->user = $user;
        $this->coordinator = $coordinator;
        $this->password = $password;
    }

    /**
     * Build the message.
     *
     * @return $this
     */
    public function build()
    {
        return $this->markdown('emails.auth.register_supervisor', [
            'user' =>  $this->user,
            'coordinator' => $this->coordinator,
            'password' => $this->password,
            'profile_completion_url' =>  $_SERVER['HTTP_HOST'] . "/profile/complete"
        ]);
    }
}
