@protocol Printing: someParent
-(void) print;
@end

@interface Fraction: NSObject <Printing, NSCopying> {
    int numerator;
    int denominator;
}
@end

@"blah\8" @"a\222sd\d" @"\faw\"\? \' \4 n\\" @"\56"
@"\xSF42"

-(NSDecimalNumber*)addCount:(id)addObject{

return [count decimalNumberByAdding:addObject.count];

}

  NS_DURING  NS_HANDLER NS_ENDHANDLER

@try {
   if (argc > 1)    {
    @throw [NSException exceptionWithName:@"Throwing a test exception" reason:@"Testing the @throw directive." userInfo:nil];
   }
} 
@catch (id theException) {
    NSLog(@"%@", theException);
    result = 1  ;
} 
@finally {
    NSLog(@"This always happens.");
    result += 2 ;
}

    @synchronized(lock) {
        NSLog(@"Hello World");
    }

struct { @defs( NSObject) }

char *enc1 = @encode(int);

         IBOutlet|IBAction|BOOL|SEL|id|unichar|IMP|Class 


 @class @protocol

@public
  // instance variables
@package
  // instance variables
@protected
  // instance variables
@private
  // instance variables

  YES NO Nil nil
NSApp()
NSRectToCGRect (Protocol ProtocolFromString:"NSTableViewDelegate"))

[SPPoint pointFromCGPoint:self.position]

NSRoundDownToMultipleOfPageSize

#import <stdio.h>

int main( int argc, const char *argv[] ) {
    printf( "hello world\n" );
    return 0;
}

NSChangeSpelling

@"0 != SUBQUERY(image, $x, 0 != SUBQUERY($x.bookmarkItems, $y, $y.@count == 0).@count).@count"

@selector(lowercaseString) @selector(uppercaseString:)

NSFetchRequest *localRequest = [[NSFetchRequest alloc] init];  
localRequest.entity = [NSEntityDescription entityForName:@"VNSource" inManagedObjectContext:context];  
localRequest.sortDescriptors = [NSArray arrayWithObject:[NSSortDescriptor sortDescriptorWithKey:@"resolution" ascending:YES]];  
NSPredicate *predicate = [NSPredicate predicateWithFormat:@"0 != SUBQUERY(image, $x, 0 != SUBQUERY($x.bookmarkItems, $y, $y.@count == 0).@count).@count"];
[NSPredicate predicateWithFormat:]
NSString *predicateString = [NSString stringWithFormat:@"SELF beginsWith[cd] %@", searchString];
NSPredicate *pred = [NSPredicate predicateWithFormat:predicateString];
NSArray *filteredKeys = [[myMutableDictionary allKeys] filteredArrayUsingPredicate:pred]; 

localRequest.predicate = [NSPredicate predicateWithFormat:@"whichChart = %@" argumentArray: listChartToDownload];
localRequest.fetchBatchSize = 100;
arrayRequest    = [context  executeFetchRequest:localRequest error:&error1];

[localRequest   release];

#ifndef Nil
#define Nil __DARWIN_NULL   /* id of Nil class */
#endif

@implementation MyObject
- (unsigned int)areaOfWidth:(unsigned int)width
                height:(unsigned int)height
{
  return width*height;
}
@end
