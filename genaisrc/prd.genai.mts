script({
    title: "Pull Request Descriptor",
    description: "Generate a pull request description from the git diff",
    temperature: 0.5,
    systemSafety: false,
    cache: true
});
const maxTokens = 7000;
const defaultBranch = await git.defaultBranch()
const branch = await git.branch();
if (branch === defaultBranch) cancel("you are already on the default branch");

// compute diff in chunks to avoid hitting context window size
const changes = await git.diff({
    base: defaultBranch,
});
const chunks = await tokenizers.chunk(changes, { chunkSize: maxTokens, chunkOverlap: 100 })
console.log(`Found ${chunks.length} chunks of changes`);
const summaries = []
for (const chunk of chunks) {
    const { text: summary, error } = await runPrompt(ctx => {
        if (summaries.length)
            ctx.def("PREVIOUS_SUMMARIES", summaries.join("\n"), { flex: 1 });
        ctx.def("GIT_DIFF", chunk, { flex: 5 })
        ctx.$`You are an expert code reviewer with great English technical writing skills and also an accomplished Go (golang) developer.

Your task is to generate a summary in a chunk of the changes in <GIT_DIFF> for a pull request in a way that a software engineer will understand.
This description will be used as the pull request description.

This summary will be concatenated with previous summaries to form the final description and will be processed by a language model.

${summaries.length ? `The previous summaries are <PREVIOUS_SUMMARIES>` : ""}
`
    }, { label: `summarizing chunk`, responseType: "text", systemSafety: true, system: [], model: "small", flexTokens: maxTokens, cache: true })
    if (error) {
        cancel(`error summarizing chunk: ${error.message}`);
    }
    summaries.push(summary)
}

def("GIT_DIFF", summaries.join("\n"), {
    maxTokens,
});

// task
$`## Task

You are an expert code reviewer with great English technical writing skills and also an accomplished Go (golang) developer.

Your task is to generate a high level summary of the changes in <GIT_DIFF> for a pull request in a way that a software engineer will understand.
This description will be used as the pull request description.

## Instructions

- generate a descriptive title for the overall changes of the pull request, not "summary". Make it fun.
- do NOT explain that GIT_DIFF displays changes in the codebase
- try to extract the intent of the changes, don't focus on the details
- use bullet points to list the changes
- use emojis to make the description more engaging
- focus on the most important changes
- do not try to fix issues, only describe the changes
- ignore comments about imports (like added, remove, changed, etc.)
`;
