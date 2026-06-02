package masker 
 
import "context" 
 
type Producer interface { 
    Produce(ctx context.Context) ([]string, error) 
} 
 
type Presenter interface { 
    Present(ctx context.Context, data []string) error 
} 
 
type Masker interface { 
    Mask(ctx context.Context, line string) string 
} 
