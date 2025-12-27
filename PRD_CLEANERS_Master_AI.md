# Ask‑Me Cleaners – Development PRD (Vercel Edition)

## 1. 목적 (Why)

이 문서는 **Ask‑Me Cleaners Master AI**를 **빠르고 안정적으로 MVP → SaaS**로 구현하기 위한 **개발자용 PRD**이다.  
배포 타깃은 **Vercel**이며, 초기 목표는 **최소한의 복잡성 + 확장 가능한 구조**다.

---

## 2. 핵심 결정 요약 (TL;DR)

- **Frontend / API**: Next.js (App Router) – Vercel Native
- **LLM**: OpenRouter (GPT‑4.1 계열)
- **Vector DB**: Pinecone (Knowledge / Partner 분리)
- **Auth**: 없음 (MVP), 추후 Clerk/Auth.js
- **DB**: 없음 (MVP), 파트너 메타는 Pinecone로 시작
- **Infra 철학**: “서버 관리 X, 코드만 관리 O”

---

## 3. 시스템 아키텍처 (단순화)

```
Client (Web / PWA)
  ↓
Next.js App Router
  ├─ /api/chat  (Server Action / Route Handler)
  │    ↓
  │  RAG Orchestrator
  │    ├─ Pinecone (Laundry Knowledge)
  │    ├─ Pinecone (Partner Cleaners)
  │    └─ OpenRouter LLM
  ↓
Response (Answer + Recommended Shops)
```

> ❗ Go 백엔드는 **MVP 범위에서 제거**  
> → 복잡도 감소, 배포/운영 비용 최소화

---

## 4. 기술 스택 (확정)

### Frontend / Backend (Unified)

- **Next.js 20+ (App Router)**
- TypeScript
- Server Actions / Route Handlers
- Edge Runtime (가능한 범위)

### AI / Data

- OpenRouter API
- Pinecone Assistant / Index
- Embeddings: text-embedding-3-large (or equivalent)

### 배포

- Vercel (Preview / Production)
- Environment Variables (Vercel Dashboard)

---

## 5. 기능 명세 (MVP Scope)

### 5.1 세탁 장인 Q&A (RAG)

**입력**

```json
{
  "message": "실크 블라우스에 와인 얼룩",
  "zipcode": "90005"
}
```

**처리 흐름**

1. 사용자 질문 수신
2. Laundry Knowledge Index 검색
3. 성공 확률 / 위험도 계산
4. 조건 충족 시 Partner Index 검색
5. 장인 캐릭터 프롬프트로 LLM 호출

**출력**

```json
{
  "answer": "...",
  "success_rate": "30~40%",
  "risk_level": "high",
  "recommended_shops": []
}
```

---

### 5.2 파트너 세탁소 추천 로직

**추천 트리거**

- success_rate < 60%
- fabric ∈ [silk, cashmere, leather]
- stain_age > 48h
- 질문에 “맡기면” 포함

**우선순위**

1. Pinecone subscription_status = active
2. zipcode match
3. specialty match

---

## 6. Pinecone 데이터 설계

### Index A – Laundry Knowledge

```json
{
  "id": "stain_wine_silk",
  "metadata": {
    "stain_type": "wine",
    "fabric": "silk",
    "success_rate": 0.35,
    "risk": "high"
  }
}
```

### Index B – Partner Cleaners

```json
{
  "id": "abc_cleaners_90005",
  "metadata": {
    "shop_name": "ABC Cleaners",
    "zipcode": "90005",
    "subscription": "active",
    "specialty": ["silk", "luxury"]
  }
}
```

---

## 7. API 설계 (Next.js Route Handler)

### POST /api/chat

- Runtime: nodejs (초기 안정성 우선)
- Timeout: 10s

---

## 8. 환경 변수 (Vercel)

```bash
OPENROUTER_API_KEY=sk-or-...
OPENROUTER_MODEL=openai/gpt-4.1
PINECONE_API_KEY=...
PINECONE_INDEX_LAUNDRY=laundry-knowledge
PINECONE_INDEX_PARTNER=partner-cleaners
```

---

## 9. 금지 규칙 (Hard Rules)

- 100% 성공 표현 금지
- 의료/법률 조언 금지
- “책임 없음” 문구 자동 삽입
- 파트너 아닌 세탁소 추천 금지

---

## 10. MVP 이후 로드맵

### Phase 2

- 이미지 업로드 → 얼룩 분류
- 성공 확률 UI 시각화
- 파트너 대시보드

### Phase 3

- 예약 / 픽업 연동
- 구독 결제 (Stripe)
- 지역 광고 슬롯

---

## 11. 개발 원칙 (중요)

- **코드는 짧게, 로직은 명확하게**
- DB 늘리기 전에 Pinecone로 해결
- “지금 필요 없는 확장”은 금지

---

## 12. Definition of Done (MVP)

- [ ] 질문 → 답변 3초 이내
- [ ] 파트너 추천 정확도 > 80%
- [ ] Vercel Preview/Prod 동일 동작
- [ ] 운영자 개입 없이 동작

---

이 문서는 **개발 중 유일한 기준 문서(Single Source of Truth)**로 사용한다.
