func TestServiceRunPresenterError(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    prod := &mockProducer{data: []string{"a1"}, err: nil}
    pres := &mockPresenter{err: errors.New("presenter error")}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)
    
    err := svc.Run(context.Background())
    
    assert.Error(t, err)
    assert.EqualError(t, err, "presenter error")
}