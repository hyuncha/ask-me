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
  const apiKey = process.env.OPENROUTER_API_KEY?.trim();
  const model = process.env.OPENROUTER_MODEL || 'openai/gpt-4o-mini';

  // 환경 변수 디버깅 - 실제 값 앞 20자 출력 (디버깅용)
  console.log('=== OPENROUTER DEBUG START ===');
  console.log('Raw API Key (first 20 chars):', apiKey?.substring(0, 20) || 'EMPTY');
  console.log('API Key length:', apiKey?.length || 0);
  console.log('Starts with sk-or-:', apiKey?.startsWith('sk-or-'));
  console.log('=== OPENROUTER DEBUG END ===');

  if (!apiKey) {
    throw new Error('OPENROUTER_API_KEY_MISSING');
  }

  // API 키 형식 검증 제거 - 일단 API 호출 시도
  // if (!apiKey.startsWith('sk-or-')) {
  //   console.error('[OpenRouter] Invalid API key format. Expected sk-or-...');
  //   throw new Error('OPENROUTER_ERROR:401:API 키 형식이 올바르지 않습니다. sk-or-로 시작해야 합니다.');
  // }

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
    const errorText = await response.text();
    console.error('OpenRouter raw error:', {
      status: response.status,
      statusText: response.statusText,
      body: errorText,
    });

    // 에러 유형별 구체적 메시지
    let userMessage = '';
    if (response.status === 401) {
      userMessage = 'API 키가 유효하지 않습니다. OpenRouter API 키를 확인해주세요.';
    } else if (response.status === 402) {
      userMessage = 'OpenRouter 크레딧이 부족합니다. 크레딧을 충전해주세요.';
    } else if (response.status === 429) {
      userMessage = '요청 한도를 초과했습니다. 잠시 후 다시 시도해주세요.';
    } else if (response.status === 400) {
      userMessage = '잘못된 요청입니다. 모델명을 확인해주세요.';
    } else {
      userMessage = `OpenRouter 오류가 발생했습니다 (${response.status})`;
    }

    throw new Error(`OPENROUTER_ERROR:${response.status}:${userMessage}`);
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
