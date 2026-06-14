interface IconProps {
  className?: string
}

export function StrawberryIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      {/* Тело клубники */}
      <path
        d="M60 108C42 92 27 76 27 53C27 34 41 27 55 36C57 37 58 38 60 40C62 38 63 37 65 36C79 27 93 34 93 53C93 76 78 92 60 108Z"
        fill="currentColor"
      />
      {/* Левый лист */}
      <path d="M60 34C55 25 45 20 36 20C43 24 45 30 45 36C39 32 31 32 24 36C34 38 40 43 44 49C48 40 54 36 60 34Z" fill="currentColor" opacity="0.6" />
      {/* Правый лист */}
      <path d="M60 34C65 25 75 20 84 20C77 24 75 30 75 36C81 32 89 32 96 36C86 38 80 43 76 49C72 40 66 36 60 34Z" fill="currentColor" opacity="0.6" />
      {/* Семечки */}
      {[
        [48, 52], [62, 55], [75, 51], [43, 68], [59, 72], [78, 70], [52, 88], [68, 90],
      ].map(([cx, cy], i) => (
        <ellipse key={i} cx={cx} cy={cy} rx="2.5" ry="4" fill="currentColor" opacity="0.5" transform={`rotate(20 ${cx} ${cy})`} />
      ))}
    </svg>
  )
}

export function BrainIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      {/* Основная форма мозга */}
      <path
        d="M37 83C26 80 20 70 22 59C14 50 19 36 31 35C34 25 48 21 57 29C64 20 80 23 83 35C96 36 102 49 96 60C103 72 94 86 80 86H42C40 86 38 85 37 83Z"
        fill="currentColor"
        opacity="0.15"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Центральная борозда */}
      <path d="M58 30V87" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      {/* Левое полушарие — извилины */}
      <path d="M50 36C42 35 35 40 34 48C27 50 25 61 32 66C30 75 37 81 46 79" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      {/* Правое полушарие — извилины */}
      <path d="M67 36C75 34 82 39 83 47C91 49 92 60 86 65C90 74 83 82 73 79" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      {/* Детали левого полушария */}
      <path d="M40 55C47 58 51 55 54 50" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      {/* Детали правого полушария */}
      <path d="M66 51C70 58 78 58 84 54" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      <path d="M45 69C50 70 54 68 58 64" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      <path d="M66 65C71 70 77 71 82 68" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}

export function CardsIcon({ className }: IconProps) {
  const cards = [
    { x: 24, y: 33, rotate: -22, rank: "A", suit: "♥" },
    { x: 33, y: 27, rotate: -11, rank: "K", suit: "♠" },
    { x: 42, y: 24, rotate: 0, rank: "Q", suit: "♦" },
    { x: 51, y: 27, rotate: 11, rank: "J", suit: "♣" },
    { x: 60, y: 33, rotate: 22, rank: "10", suit: "♥" },
  ]
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      <g transform="translate(0 4)">
        {cards.map((card, i) => (
          <g key={i} transform={`rotate(${card.rotate} 60 92)`}>
            {/* Рамка карты */}
            <rect x={card.x} y={card.y} width="36" height="58" rx="6" fill="currentColor" opacity="0.08" stroke="currentColor" strokeWidth="2.5" />
            {/* Ранг */}
            <text x={card.x + 8} y={card.y + 18} fontSize="13" fontWeight="700" fill="currentColor" fontFamily="Arial, sans-serif">{card.rank}</text>
            {/* Масть */}
            <text x={card.x + 18} y={card.y + 39} textAnchor="middle" fontSize="20" fill="currentColor" fontFamily="Arial, sans-serif">{card.suit}</text>
          </g>
        ))}
      </g>
    </svg>
  )
}

export function ClosedBookIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      {/* Основная обложка */}
      <path
        d="M34 23H82C88 23 92 27 92 33V90C92 95 88 99 83 99H34C29 99 26 96 26 91V31C26 26 29 23 34 23Z"
        fill="currentColor"
        opacity="0.18"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinejoin="round"
      />
      {/* Страницы снизу */}
      <path
        d="M36 82H91V93C91 97 88 99 84 99H36C30 99 26 96 26 91C26 86 30 82 36 82Z"
        fill="currentColor"
        opacity="0.08"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      {/* Линии страниц */}
      <line x1="36" y1="86" x2="88" y2="86" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.3" />
      <line x1="36" y1="91" x2="86" y2="91" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.3" />
      {/* Корешок */}
      <path d="M39 23V82" stroke="currentColor" strokeWidth="4" strokeLinecap="round" opacity="0.4" />
      {/* Этикетка на обложке */}
      <rect x="50" y="38" width="27" height="24" rx="4" fill="currentColor" opacity="0.1" stroke="currentColor" strokeWidth="2" />
      {/* Текст на этикетке */}
      <line x1="56" y1="48" x2="71" y2="48" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.6" />
      <line x1="56" y1="54" x2="68" y2="54" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.6" />
      {/* Закладка-ленточка */}
      <path d="M80 24V62L73 56L66 62V24H80Z" fill="currentColor" opacity="0.25" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" />
    </svg>
  )
}

export function OpenBookIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      {/* Левая обложка */}
      <path
        d="M60 38C49 29 35 28 22 33V88C36 83 49 84 60 94V38Z"
        fill="currentColor"
        opacity="0.15"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinejoin="round"
      />
      {/* Правая обложка */}
      <path
        d="M60 38C71 29 85 28 98 33V88C84 83 71 84 60 94V38Z"
        fill="currentColor"
        opacity="0.15"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinejoin="round"
      />
      {/* Левая страница */}
      <path
        d="M60 33C49 25 36 24 25 29V82C38 78 50 80 60 89V33Z"
        fill="currentColor"
        opacity="0.06"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      {/* Правая страница */}
      <path
        d="M60 33C71 25 84 24 95 29V82C82 78 70 80 60 89V33Z"
        fill="currentColor"
        opacity="0.06"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      {/* Центральный переплёт */}
      <line x1="60" y1="33" x2="60" y2="91" stroke="currentColor" strokeWidth="3" strokeLinecap="round" opacity="0.5" />
      {/* Текстовые линии — левая страница */}
      <line x1="35" y1="42" x2="55" y2="45" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
      <line x1="35" y1="53" x2="55" y2="57" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
      <line x1="35" y1="64" x2="55" y2="68" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
      {/* Текстовые линии — правая страница */}
      <line x1="85" y1="42" x2="65" y2="45" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
      <line x1="85" y1="53" x2="65" y2="57" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
      <line x1="85" y1="64" x2="65" y2="68" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" opacity="0.35" />
    </svg>
  )
}

export function MysticOrbIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 120 120" fill="none" role="img">
      {/* Шар */}
      <circle cx="60" cy="50" r="34" fill="currentColor" opacity="0.12" stroke="currentColor" strokeWidth="3" />
      {/* Подставка */}
      <path d="M43 89H77L84 106H36L43 89Z" fill="currentColor" opacity="0.18" stroke="currentColor" strokeWidth="3" strokeLinejoin="round" />
      {/* Основание подставки */}
      <ellipse cx="60" cy="89" rx="24" ry="8" fill="currentColor" opacity="0.18" stroke="currentColor" strokeWidth="3" />
      {/* Внутренние дуги шара */}
      <path d="M43 47C52 38 68 38 77 47" stroke="currentColor" strokeWidth="3" strokeLinecap="round" opacity="0.55" />
      <path d="M50 61C57 66 66 66 73 61" stroke="currentColor" strokeWidth="3" strokeLinecap="round" opacity="0.35" />
      {/* Блики */}
      <circle cx="49" cy="38" r="5" fill="currentColor" opacity="0.25" />
      <circle cx="39" cy="56" r="2.5" fill="currentColor" opacity="0.2" />
      <circle cx="78" cy="35" r="2" fill="currentColor" opacity="0.2" />
      {/* Звёздочка */}
      <path d="M60 25L63 33L71 36L63 39L60 47L57 39L49 36L57 33L60 25Z" fill="currentColor" opacity="0.55" />
    </svg>
  )
}