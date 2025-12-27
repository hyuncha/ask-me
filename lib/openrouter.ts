import { OpenRouterMessage, OpenRouterResponse } from './types';

const OPENROUTER_API_URL = 'https://openrouter.ai/api/v1/chat/completions';

// 세탁 장인 시스템 프롬프트
export const LAUNDRY_MASTER_PROMPT = `너는 30년 경력의 세탁 장인이다. 세탁, 얼룩 제거, 의류 소재 관리에 대해 전문적이고 솔직하게 답변한다.

## 응답 규칙
1. 될 수 있는 것과 안 되는 것을 명확히 구분해서 말해라
2. 성공 확률을 구체적으로 설명해라 (예: "이 경우 성공률은 30~40% 정도입니다")
3. 집에서 시도할 때의 위험성을 반드시 경고해라
4. 책임 회피 없이 현실적인 조언을 해라
5. 100% 성공을 보장하는 표현은 절대 사용하지 마라

## 파트너 세탁소 추천 조건
다음 조건 중 하나라도 해당되면 전문 세탁소를 추천해라:
- 성공 확률이 60% 미만인 경우
- 고급 소재인 경우 (실크, 캐시미어, 가죽, 울, 린넨 등)
- 얼룩 발생 후 48시간이 초과된 경우
- 고객이 "맡기면 나을까요?" 또는 유사한 질문을 한 경우

## 응답 형식
응답 마지막에 반드시 다음 JSON 블록을 포함해라 (파싱용):
\`\`\`json
{
  "success_rate": "예상 성공률 (예: 30~40%)",
  "risk_level": "low|medium|high",
  "recommend_shop": true|false
}
\`\`\`

## 말투
- 친근하지만 전문가다운 말투를 사용해라
- "이건 집에서 건드리면 거의 망가집니다" 같은 직설적 표현을 써라
- 경험에서 우러나온 조언처럼 말해라`;

export async function sendMessage(
  userMessage: string,
  context?: string
): Promise<{ content: string; raw: OpenRouterResponse }> {
  const apiKey = process.env.OPENROUTER_API_KEY;
  const model = process.env.OPENROUTER_MODEL || 'openai/gpt-4o-mini';

  // 환경 변수 디버깅 로그 (민감값 마스킹)
  console.log('OPENROUTER_API_KEY set:', !!apiKey);
  console.log('OPENROUTER_MODEL:', model);

  if (!apiKey) {
    throw new Error('OPENROUTER_API_KEY_MISSING');
  }

  const systemPrompt = context
    ? `${LAUNDRY_MASTER_PROMPT}\n\n## 관련 세탁 지식 (검색 결과):\n${context}`
    : LAUNDRY_MASTER_PROMPT;

  const messages: OpenRouterMessage[] = [
    { role: 'system', content: systemPrompt },
    { role: 'user', content: userMessage },
  ];

  const response = await fetch(OPENROUTER_API_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
      'HTTP-Referer': 'https://ask-me-cleaners.vercel.app',
      'X-Title': 'Ask-Me Cleaners',
    },
    body: JSON.stringify({
      model,
      messages,
      temperature: 0.7,
      max_tokens: 1500,
    }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`OpenRouter API error: ${response.status} - ${error}`);
  }

  const data: OpenRouterResponse = await response.json();
  const content = data.choices[0]?.message?.content || '';

  return { content, raw: data };
}

// 응답에서 JSON 메타데이터 파싱
export function parseResponseMetadata(content: string): {
  success_rate?: string;
  risk_level?: 'low' | 'medium' | 'high';
  recommend_shop?: boolean;
  cleanContent: string;
} {
  const jsonMatch = content.match(/```json\s*([\s\S]*?)\s*```/);

  if (!jsonMatch) {
    return { cleanContent: content };
  }

  try {
    const metadata = JSON.parse(jsonMatch[1]);
    const cleanContent = content.replace(/```json[\s\S]*?```/, '').trim();

    return {
      success_rate: metadata.success_rate,
      risk_level: metadata.risk_level,
      recommend_shop: metadata.recommend_shop,
      cleanContent,
    };
  } catch {
    return { cleanContent: content };
  }
}
