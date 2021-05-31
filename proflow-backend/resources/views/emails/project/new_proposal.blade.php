@component('mail::message')
# A new project has been proposed

Dear supervisor {{$supervisor->name}},

{{$student->name}}, recently {{$action->new_proposal ? "proposed": "reproposed"}} a project on the topic, {{ $project->name }}.

Review this proposal [here]() or click the button below.

@component('mail::button', ['url' => ''])
Review proposal
@endcomponent

Thanks,<br>
{{ config('app.name') }}
@endcomponent
