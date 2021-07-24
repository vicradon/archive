<?php

namespace App\Notifications;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Notifications\Messages\MailMessage;
use Illuminate\Notifications\Notification;

class ProposalApproved extends Notification
{
    use Queueable;

    /**
     * Create a new notification instance.
     *
     * @return void
     */
    public function __construct($proposal)
    {
        $this->proposal = $proposal;
    }

    /**
     * Get the notification's delivery channels.
     *
     * @param  mixed  $notifiable
     * @return array
     */
    public function via($notifiable)
    {
        return ['mail'];
    }

    /**
     * Get the mail representation of the notification.
     *php
     * @param  mixed  $notifiable
     * @return \Illuminate\Notifications\Messages\MailMessage
     */
    public function toMail($notifiable)
    {
        $proposal_name = $this->proposal->name;
        $frontend_root = config('app.env') === 'local' ?
            config('app.frontend_local_root_url') :
            config('app.frontend_remote_root_url');

        $url = "{$frontend_root}/student/project/upload";

        return (new MailMessage)
            ->subject('Proposal Approved')
            ->line("Your proposal with title: ${proposal_name} has been approved.")
            ->line("You can start uploading your project's chapters")
            ->action('Notification Action', url($url))
            ->line('Thank you for using Proflow!');
    }

    /**
     * Get the array representation of the notification.
     *
     * @param  mixed  $notifiable
     * @return array
     */
    public function toArray($notifiable)
    {
        return [
            //
        ];
    }
}
