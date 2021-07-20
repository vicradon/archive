<?php

namespace App\Notifications;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Notifications\Messages\MailMessage;
use Illuminate\Notifications\Notification;

class NewProjectUpload extends Notification implements ShouldQueue
{
    use Queueable;

    /**
     * Create a new notification instance.
     *
     * @return void
     */
    public function __construct($student_user)
    {
        $this->student_user = $student_user;
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
     *
     * @param  mixed  $notifiable
     * @return \Illuminate\Notifications\Messages\MailMessage
     */
    public function toMail($notifiable)
    {
        $student_user_name = $this->student_user->name;
        $student_user_id = $this->student_user->id;

        $frontend_root = config('app.env') === 'local' ?
            config('app.frontend_local_root_url') :
            config('app.frontend_remote_root_url');
        $url = "{$frontend_root}/supervisor/students/{$student_user_id}";

        return (new MailMessage)
            ->subject('New Project Uploaded')
            ->line("A new project PDF has been uploaded by ${student_user_name}.")
            ->line("Please review this PDF by leaving appropriate feedback where necessary.")
            ->action('Notification Action', url($url))
            ->line('Thank you for using Proflow.');
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
