package util

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

func ConsultChatGPT(content []string, requirements string) (string, error) {
	client := openai.NewClient()

	prompt := fmt.Sprintf(`
You are a software design analyst. Use the CLEAR framework (Context, Lens, Expectations, Analysis, Recommendations) to evaluate how well the following code meets the specified system design requirements.

--- SYSTEM DESIGN REQUIREMENTS ---
%s

--- CODE CONTENT ---
%s

Format the following technical analysis into a structured, professional document. Use clear section headers, bullet points where appropriate, and improve readability while preserving all original meaning and detail. Group related findings and recommendations together under logical headings such as "Security Gaps", "Feature Limitations", "Recommendations", etc. The final result should resemble a security audit report or product evaluation summary that’s suitable for stakeholders or developers.



**Context**: What is the context of this system and the intent behind the code?
**Lens**: What criteria or perspective are you using to evaluate the code against the requirements?
**Expectations**: What does the system design require in terms of architecture, logic, or design?
**Analysis**: How does the code satisfy (or fail to satisfy) each requirement? Provide reasoning and examples.
**Recommendations**: What can be improved in the code to better align with the design requirements?
`, requirements, content)

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{

			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return "Error getting ChatGPT response", nil
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
