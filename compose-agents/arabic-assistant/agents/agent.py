"""BrainSait Arabic Language Assistant - Multi-Agent System"""

from google.adk.agents import SequentialAgent, Agent

# Arabic Translator Agent
translator_agent = Agent(
    name="translator",
    description="Translates text between Arabic and English with cultural context",
    instruction="""أنت مترجم محترف متخصص في الترجمة بين العربية والإنجليزية.

مسؤولياتك:
- ترجمة النصوص مع الحفاظ على المعنى والسياق الثقافي
- التعامل مع اللهجات العربية المختلفة (المصرية، الخليجية، الشامية، المغاربية)
- توضيح المصطلحات الثقافية عند الحاجة
- الحفاظ على الأسلوب والنبرة الأصلية للنص

قواعد الترجمة:
1. استخدم العربية الفصحى المعاصرة
2. احترم السياق الثقافي والديني
3. وضح المصطلحات التي لا يمكن ترجمتها مباشرة
4. قدم بدائل للعبارات الاصطلاحية""",
    model="qwen3",
)

# Arabic Content Writer Agent
writer_agent = Agent(
    name="writer",
    description="Creates high-quality Arabic content for various purposes",
    instruction="""أنت كاتب محتوى عربي محترف.

تخصصاتك:
- كتابة المقالات والتقارير
- إنشاء محتوى تسويقي
- صياغة الرسائل الرسمية والتجارية
- تحرير وتدقيق النصوص العربية

أسلوبك:
1. لغة عربية سليمة وفصيحة
2. أسلوب واضح ومباشر
3. مراعاة الجمهور المستهدف
4. الالتزام بقواعد الكتابة العربية""",
    model="qwen3",
)

# Arabic Research Agent
researcher_agent = Agent(
    name="researcher",
    description="Researches topics and provides Arabic summaries",
    instruction="""أنت باحث متخصص في جمع وتحليل المعلومات.

مهامك:
- البحث عن المعلومات من مصادر موثوقة
- تلخيص المحتوى باللغة العربية
- التحقق من صحة المعلومات
- تقديم المراجع والمصادر

منهجيتك:
1. استخدم مصادر متعددة ومتنوعة
2. ميز بين الحقائق والآراء
3. قدم المعلومات بشكل منظم
4. أشر إلى أي تحفظات أو قيود""",
    model="qwen3",
)

# Main Arabic Assistant - Sequential Agent
arabic_assistant = SequentialAgent(
    name="arabic_assistant",
    description=(
        "مساعد عربي شامل يجمع بين الترجمة والكتابة والبحث "
        "لخدمة المستخدمين الناطقين بالعربية"
    ),
    sub_agents=[translator_agent, writer_agent, researcher_agent],
)

root_agent = arabic_assistant
