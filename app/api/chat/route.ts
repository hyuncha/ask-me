import { NextRequest, NextResponse } from 'next/server';
import { processChat } from '@/lib/rag';
import { ChatRequest } from '@/lib/types';

export const runtime = 'nodejs';
export const maxDuration = 30; // 30초 타임아웃

export async function POST(request: NextRequest) {
  // 환경 변수 디버깅 (민감값 마스킹)
  console.log('ENV check - OPENROUTER_API_KEY set:', !!process.env.OPENROUTER_API_KEY);

  try {
    const body: ChatRequest = await request.json();

    if (!body.message || typeof body.message !== 'string') {
      return NextResponse.json(
        { error: '메시지를 입력해주세요.' },
        { status: 400 }
      );
    }

    if (body.message.length > 1000) {
      return NextResponse.json(
        { error: '메시지가 너무 깁니다. 1000자 이내로 입력해주세요.' },
        { status: 400 }
      );
    }

    const response = await processChat(body.message, body.zipcode);

    return NextResponse.json(response);
  } catch (error) {
    console.error('Chat API error:', error);

    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';

    // OpenRouter API 키 미설정 - 400으로 반환
    if (errorMessage.includes('OPENROUTER_API_KEY_MISSING')) {
      return NextResponse.json(
        { error: 'OPENROUTER_API_KEY 환경 변수가 설정되지 않았습니다. Vercel Dashboard에서 설정해주세요.' },
        { status: 400 }
      );
    }

    // OpenRouter API 에러 상세 처리 (OPENROUTER_ERROR:status:message 형식)
    if (errorMessage.startsWith('OPENROUTER_ERROR:')) {
      const parts = errorMessage.split(':');
      const statusCode = parseInt(parts[1]) || 502;
      const userMessage = parts.slice(2).join(':') || 'OpenRouter API 오류';

      console.error('OpenRouter API error details:', { statusCode, userMessage });

      // 401(인증오류), 402(크레딧부족)은 400으로 반환 (클라이언트가 수정해야 할 문제)
      // 429(제한초과), 500+(서버오류)은 502로 반환
      return NextResponse.json(
        { error: userMessage },
        { status: statusCode === 401 || statusCode === 402 ? 400 : 502 }
      );
    }

    return NextResponse.json(
      { error: '죄송합니다. 잠시 후 다시 시도해주세요.' },
      { status: 500 }
    );
  }
}
