import os, httpx
from pprint import pprint
import uvicorn, json
from uuid import uuid4
from fastapi import FastAPI, Request, status, HTTPException, BackgroundTasks
from fastapi.responses import HTMLResponse
from dotenv import load_dotenv
from services import AgentService
import lxml

from typing import Union

from a2a.types import (
    AgentCard,
    AgentCapabilities,
    AgentSkill,
    AgentProvider,
    SendMessageRequest,
    GetTaskRequest
)
import a2a.types as a2a_types



load_dotenv()

TELEX_API_KEY = os.getenv('TELEX_API_KEY')
TELEX_API_URL = os.getenv('TELEX_API_URL')

app = FastAPI()


@app.get("/", response_class=HTMLResponse)
def read_root():
    return '<p style="font-size:30px">SEO agent</p>'


@app.get("/.well-known/agent.json")
def get_agent_card(request: Request):
    external_base = request.headers.get("x-external-base-url", "")
    current_base_url = str(request.base_url).rstrip("/") + external_base

    capabilities = AgentCapabilities(pushNotifications=True)

    skills = AgentSkill(
        id= "seo-audit",
        name= "SEO Auditing for webpages",
        description= "Audits Webpages for SEO issues",
        inputModes= ["text"],
        outputModes= ["text"],
        tags=["seo", "seo-audit"]
    )

    provider = AgentProvider(
       organization="Telex", url="https://telex.im"
    )

    agent_card = AgentCard(
        name='SEO Audit Agent - Gitpod-Test1',
        description='Audits Webpages for SEO issues',
        url=current_base_url,
        version='1.0.1',
        defaultInputModes=["text/plain"],
        defaultOutputModes=["text/plain"],
        capabilities=capabilities,
        skills=[skills],
        provider=provider,
        documentationUrl=f"{current_base_url}/docs"
    )

    card = AgentCard.model_dump(agent_card)

    return card


async def handle_task(message:str, request_id, user_id:str, task_id: str, webhook_url: str, api_key: str, context_id: str):
  links = await AgentService.discover_links(start_url=message, max_pages=3)
  data = await AgentService.audit_multiple_pages_with_ai(links=links, api_key=api_key)
  pprint(data)

  return "stopping..."

  parts = a2a_types.TextPart(text=data)

  message = a2a_types.Message(messageId=uuid4().hex, role=a2a_types.Role.agent, parts=[parts])

  artifacts = a2a_types.Artifact(artifactId=uuid4().hex, parts=[parts])

  task = a2a_types.Task(
    id = task_id,
    contextId= context_id,
    status =  a2a_types.TaskStatus(
      state=a2a_types.TaskState.completed, 
      message=a2a_types.Message(messageId=uuid4().hex, role=a2a_types.Role.agent, parts=[a2a_types.TextPart(text="Success!")])
    ),
    artifacts = [artifacts]
  )

  webhook_response = a2a_types.SendMessageSuccessResponse(
      id=request_id,
      result=task
  )

  pprint(webhook_response.model_dump(exclude_none=True))


  async with httpx.AsyncClient() as client:
    headers = {"X-TELEX-API-KEY": api_key}
    is_sent = await client.post(webhook_url, headers=headers,  json=webhook_response.model_dump(exclude_none=True, mode="json"))
    pprint(is_sent.json())

  print("background done")
  return 


async def handle_get_task_request(request: GetTaskRequest):
  pass

async def handle_send_message_request(request: SendMessageRequest, background_tasks: BackgroundTasks):
  user_id = request.params.message.metadata.get("telex_user_id")
  org_id = request.params.message.metadata.get("org_id")  
  webhook_url = request.params.configuration.pushNotificationConfig.url
  api_key = request.params.configuration.pushNotificationConfig.authentication.credentials or TELEX_API_KEY
  message_parts = request.params.message.parts

  if not message_parts:
    raise HTTPException(
      status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
      detail="Message cannot be empty."
    )

  context_id = uuid4().hex

  new_task = a2a_types.Task(
    id = uuid4().hex,
    contextId=context_id,
    status =  a2a_types.TaskStatus(
      state=a2a_types.TaskState.submitted, 
      message=a2a_types.Message(messageId=uuid4().hex, role=a2a_types.Role.agent, parts=[a2a_types.TextPart(text="your task is in progress")])
    )
  )

  combined_text_parts = ""

  for part in message_parts:
    if isinstance(part.root, a2a_types.TextPart):
      combined_text_parts += part.root.text


  if not combined_text_parts:
    raise HTTPException(
      status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
      detail="Message text cannot be empty."
    )


  background_tasks.add_task(handle_task, combined_text_parts, request.id, user_id, new_task.id, webhook_url, api_key, context_id)

  response = a2a_types.SendMessageSuccessResponse(
      result=new_task
  )

  response = response.model_dump(exclude_none=True)

  return response

@app.post("/")
async def handle_request(request: Union[SendMessageRequest, GetTaskRequest], background_tasks: BackgroundTasks):
  if isinstance(request, SendMessageRequest):
    return await handle_send_message_request(request, background_tasks)
  elif isinstance(request, GetTaskRequest):
    return await handle_get_task_request(request)
  else:
    return a2a_types.SendMessageSuccessResponse(result=a2a_types.Message(parts=[a2a_types.TextPart(text="Unhandled A2A type supplied")]))


if __name__ == "__main__":
    port = int(os.getenv("PORT", 4000))
    uvicorn.run("main:app", host="127.0.0.1", port=port, reload=True)


