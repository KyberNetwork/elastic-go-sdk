package entities

// A data provider for ticks that is backed by an in-memory array of ticks.
type TickListDataProvider struct {
	ticks []Tick
}

func NewTickListDataProvider(ticks []Tick, tickSpacing int) (*TickListDataProvider, error) {
	if err := ValidateList(ticks, tickSpacing); err != nil {
		return nil, err
	}
	return &TickListDataProvider{ticks: ticks}, nil
}

func (p *TickListDataProvider) GetNearestCurrentTick(currentTick int) (int, error) {
	return GetNearestCurrentTick(p.ticks, currentTick)
}

func (p *TickListDataProvider) TransformToMap() (map[int]TickData, map[int]LinkedListData) {
	return TransformToMap(p.ticks)
}
