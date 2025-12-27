import { ChatResponse, PartnerShop } from './types';
import { sendMessage, parseResponseMetadata } from './openrouter';
import {
  searchLaundryKnowledge,
  searchPartnerShops,
  formatKnowledgeAsContext,
} from './pinecone';

// 고급 소재 키워드
const LUXURY_FABRICS = ['실크', '캐시미어', '가죽', '울', '린넨', 'silk', 'cashmere', 'leather', 'wool', 'linen'];

// 전문 세탁소 추천 트리거 문구
const SHOP_TRIGGER_PHRASES = ['맡기', '세탁소', '전문', '의뢰', '드라이클리닝'];

// 메시지에서 고급 소재 감지
function detectLuxuryFabric(message: string): boolean {
  const lowerMessage = message.toLowerCase();
  return LUXURY_FABRICS.some((fabric) => lowerMessage.includes(fabric.toLowerCase()));
}

// 메시지에서 세탁소 추천 요청 감지
function detectShopRequest(message: string): boolean {
  return SHOP_TRIGGER_PHRASES.some((phrase) => message.includes(phrase));
}

// 파트너 추천 조건 확인
function shouldRecommendPartner(
  message: string,
  metadata: { success_rate?: string; risk_level?: string; recommend_shop?: boolean }
): boolean {
  // LLM이 추천한다고 판단한 경우
  if (metadata.recommend_shop) {
    return true;
  }

  // 고급 소재인 경우
  if (detectLuxuryFabric(message)) {
    return true;
  }

  // 사용자가 직접 세탁소를 요청한 경우
  if (detectShopRequest(message)) {
    return true;
  }

  // 성공률이 60% 미만인 경우
  if (metadata.success_rate) {
    const rateMatch = metadata.success_rate.match(/(\d+)/);
    if (rateMatch) {
      const rate = parseInt(rateMatch[1], 10);
      if (rate < 60) {
        return true;
      }
    }
  }

  // 위험도가 높은 경우
  if (metadata.risk_level === 'high') {
    return true;
  }

  return false;
}

// RAG 처리 메인 함수
export async function processChat(
  message: string,
  zipcode?: string
): Promise<ChatResponse> {
  // 1. Knowledge Index에서 관련 정보 검색
  const knowledgeResults = await searchLaundryKnowledge(message, 3);
  const context = formatKnowledgeAsContext(knowledgeResults);

  // 2. OpenRouter로 LLM 호출
  const { content } = await sendMessage(message, context || undefined);

  // 3. 응답 메타데이터 파싱
  const { success_rate, risk_level, recommend_shop, cleanContent } = parseResponseMetadata(content);

  // 4. 파트너 추천 조건 확인
  const shouldRecommend = shouldRecommendPartner(message, {
    success_rate,
    risk_level,
    recommend_shop,
  });

  // 5. 파트너 세탁소 검색 (조건 충족시)
  let recommended_shops: PartnerShop[] = [];
  if (shouldRecommend) {
    recommended_shops = await searchPartnerShops(message, zipcode, 3);
  }

  // 6. 응답 구성
  return {
    answer: cleanContent,
    success_rate,
    risk_level: risk_level as 'low' | 'medium' | 'high' | undefined,
    recommended_shops,
    disclaimer: '※ 이 조언은 참고용이며, 실제 결과는 다를 수 있습니다. 귀중한 의류는 전문 세탁소에 맡기시는 것을 권장합니다.',
  };
}
