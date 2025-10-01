import google.generativeai as genai
import structlog

logger = structlog.get_logger()

class JobProcessor:
    def __init__(self, api_key):
        genai.configure(api_key=api_key)
        self.model = genai.GenerativeModel('gemini-2.0-flash')
        logger.info("processor_initialized")
    
    def process_text(self, text):
        logger.info("processing_text", text_len=len(text))
        response = self.model.generate_content(text)
        logger.info("text_processed")
        logger.info("response", response=response)
        return response.text
    
    def process_summarize(self, text):
        logger.info("processing_summarize", text_len=len(text))
        prompt = f"Summarize this text: {text}"
        response = self.model.generate_content(prompt)
        logger.info("summarize_processed")
        return response.text
