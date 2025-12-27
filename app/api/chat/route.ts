import { NextRequest, NextResponse } from 'next/server';
import { processChat } from '@/lib/rag';
import { ChatRequest } from '@/lib/types';

export const runtime = 'nodejs';
export const maxDuration = 30; // 30초 타임아웃

export async function POST(request: NextRequest) {
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

    // OpenRouter API 키 문제
    if (errorMessage.includes('OPENROUTER_API_KEY')) {
      return NextResponse.json(
        { error: 'API 설정이 필요합니다. 관리자에게 문의하세요.' },
        { status: 500 }
      );
    }

    return NextResponse.json(
      { error: '죄송합니다. 잠시 후 다시 시도해주세요.' },
      { status: 500 }
    );
  }
}
